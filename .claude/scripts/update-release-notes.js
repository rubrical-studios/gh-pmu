#!/usr/bin/env node
/**
 * update-release-notes.js - Update GitHub release notes from CHANGELOG.md
 *
 * Parses CHANGELOG.md to extract a specific version's content and updates
 * the corresponding GitHub release with formatted release notes.
 *
 * Usage:
 *   node update-release-notes.js --version v0.9.1
 *   node update-release-notes.js --version v0.9.1 --dry-run
 *
 * Options:
 *   --version <version>   Version to update (required)
 *   --dry-run             Show what would be updated without making changes
 *   --quiet               Suppress non-output messages
 *   --help                Show this help message
 *
 * Exit codes:
 *   0 - Success
 *   1 - Error
 */

const fs = require('fs');
const path = require('path');
const gh = require('./lib/gh');
const out = require('./lib/output');

function showHelp() {
    console.log(`
update-release-notes.js - Update GitHub release notes from CHANGELOG.md

Usage:
  node update-release-notes.js --version v0.9.1
  node update-release-notes.js --version v0.9.1 --dry-run

Options:
  --version <version>   Version to update (required)
  --dry-run             Show what would be updated without making changes
  --quiet               Suppress non-output messages
  --help                Show this help message

The script:
1. Reads CHANGELOG.md and extracts the section for the specified version
2. Transforms the changelog content into GitHub release format
3. Updates the GitHub release with the new notes

Example output format:
  ## What's New

  ### Features
  - New feature description

  ### Bug Fixes
  - Fix description (#123)

  **Full Changelog**: https://github.com/owner/repo/compare/v0.9.0...v0.9.1
`);
}

/**
 * Parse CHANGELOG.md and extract content for a specific version
 * @param {string} changelogPath - Path to CHANGELOG.md
 * @param {string} version - Version to extract (e.g., "v0.9.1" or "0.9.1")
 * @returns {object} Parsed changelog section with metadata
 */
function parseChangelog(changelogPath, version) {
    const content = fs.readFileSync(changelogPath, 'utf8');
    const cleanVersion = version.replace(/^v/, '');

    // Find the section for this version
    const versionRegex = new RegExp(`^## \\[${cleanVersion}\\].*$`, 'm');
    const match = content.match(versionRegex);

    if (!match) {
        return null;
    }

    const startIndex = match.index;

    // Find the next version section or end of file
    const nextVersionRegex = /^## \[\d+\.\d+\.\d+\]/m;
    const remainingContent = content.slice(startIndex + match[0].length);
    const nextMatch = remainingContent.match(nextVersionRegex);

    let sectionContent;
    if (nextMatch) {
        sectionContent = remainingContent.slice(0, nextMatch.index).trim();
    } else {
        sectionContent = remainingContent.trim();
    }

    // Parse the header for date
    const headerMatch = match[0].match(/## \[[\d.]+\] - (\d{4}-\d{2}-\d{2})/);
    const date = headerMatch ? headerMatch[1] : null;

    // Parse sections
    const sections = {
        added: [],
        changed: [],
        deprecated: [],
        removed: [],
        fixed: [],
        security: [],
        performance: []
    };

    let currentSection = null;
    const lines = sectionContent.split('\n');

    for (const line of lines) {
        const sectionMatch = line.match(/^### (Added|Changed|Deprecated|Removed|Fixed|Security|Performance)/i);
        if (sectionMatch) {
            currentSection = sectionMatch[1].toLowerCase();
            continue;
        }

        if (currentSection && line.startsWith('- ')) {
            sections[currentSection].push(line.slice(2));
        }
    }

    return {
        version: cleanVersion,
        date,
        sections,
        rawContent: sectionContent
    };
}

/**
 * Transform changelog sections into GitHub release notes format
 * @param {object} parsed - Parsed changelog data
 * @param {string} repo - Repository in owner/repo format
 * @param {string} previousVersion - Previous version for comparison link
 * @returns {string} Formatted release notes
 */
function formatReleaseNotes(parsed, repo, previousVersion) {
    const lines = [];

    // What's New section combining Added, Changed, Performance
    const whatsNew = [
        ...parsed.sections.added,
        ...parsed.sections.changed,
        ...parsed.sections.performance
    ];

    if (whatsNew.length > 0) {
        lines.push("## What's New\n");
        for (const item of whatsNew) {
            lines.push(`- ${item}`);
        }
        lines.push('');
    }

    // Bug Fixes
    if (parsed.sections.fixed.length > 0) {
        lines.push("## Bug Fixes\n");
        for (const item of parsed.sections.fixed) {
            lines.push(`- ${item}`);
        }
        lines.push('');
    }

    // Security
    if (parsed.sections.security.length > 0) {
        lines.push("## Security\n");
        for (const item of parsed.sections.security) {
            lines.push(`- ${item}`);
        }
        lines.push('');
    }

    // Deprecated
    if (parsed.sections.deprecated.length > 0) {
        lines.push("## Deprecated\n");
        for (const item of parsed.sections.deprecated) {
            lines.push(`- ${item}`);
        }
        lines.push('');
    }

    // Removed
    if (parsed.sections.removed.length > 0) {
        lines.push("## Removed\n");
        for (const item of parsed.sections.removed) {
            lines.push(`- ${item}`);
        }
        lines.push('');
    }

    // Full changelog link
    if (repo && previousVersion) {
        const prevClean = previousVersion.replace(/^v/, '');
        lines.push(`**Full Changelog**: https://github.com/${repo}/compare/v${prevClean}...v${parsed.version}`);
    }

    return lines.join('\n').trim();
}

/**
 * Find the previous version from CHANGELOG.md
 * @param {string} changelogPath - Path to CHANGELOG.md
 * @param {string} currentVersion - Current version
 * @returns {string|null} Previous version or null
 */
function findPreviousVersion(changelogPath, currentVersion) {
    const content = fs.readFileSync(changelogPath, 'utf8');
    const cleanVersion = currentVersion.replace(/^v/, '');

    // Find all version headers
    const versionRegex = /## \[(\d+\.\d+\.\d+)\]/g;
    const versions = [];
    let match;

    while ((match = versionRegex.exec(content)) !== null) {
        versions.push(match[1]);
    }

    // Find current version index and return the next one (previous release)
    const currentIndex = versions.indexOf(cleanVersion);
    if (currentIndex >= 0 && currentIndex < versions.length - 1) {
        return versions[currentIndex + 1];
    }

    return null;
}

async function main() {
    const flags = out.parseFlags();

    if (flags.help) {
        showHelp();
        process.exit(0);
    }

    const version = out.getFlag(flags.args, '--version');
    const dryRun = flags.args.includes('--dry-run');
    const quiet = flags.args.includes('--quiet');

    if (!version) {
        out.error('Version is required. Use --version <version>');
        process.exit(1);
    }

    // Find CHANGELOG.md
    const changelogPath = path.join(process.cwd(), 'CHANGELOG.md');
    if (!fs.existsSync(changelogPath)) {
        out.error('CHANGELOG.md not found in current directory');
        process.exit(1);
    }

    // Parse changelog
    const parsed = parseChangelog(changelogPath, version);
    if (!parsed) {
        out.error(`Version ${version} not found in CHANGELOG.md`);
        process.exit(1);
    }

    if (!quiet) {
        out.info(`Found changelog entry for v${parsed.version} (${parsed.date})`);
    }

    // Get repository info
    const repo = gh.getCurrentRepo();
    if (!repo) {
        out.error('Could not determine repository. Run from a git repository.');
        process.exit(1);
    }

    // Find previous version for comparison link
    const previousVersion = findPreviousVersion(changelogPath, version);

    // Format release notes
    const releaseNotes = formatReleaseNotes(parsed, repo, previousVersion);

    if (dryRun) {
        console.log('\n--- Release Notes Preview ---\n');
        console.log(releaseNotes);
        console.log('\n--- End Preview ---\n');
        out.info('Dry run - no changes made');
        process.exit(0);
    }

    // Update the release
    const tag = version.startsWith('v') ? version : `v${version}`;

    try {
        // Write notes to temp file for proper escaping
        const tempFile = path.join(process.cwd(), 'tmp', `release-notes-${Date.now()}.md`);
        const tempDir = path.dirname(tempFile);
        if (!fs.existsSync(tempDir)) {
            fs.mkdirSync(tempDir, { recursive: true });
        }
        fs.writeFileSync(tempFile, releaseNotes);

        gh.exec(`release edit ${tag} --notes-file "${tempFile}"`);

        // Clean up temp file
        fs.unlinkSync(tempFile);

        out.success(`Updated release notes for ${tag}`);

        // Output JSON for programmatic use
        console.log(JSON.stringify({
            success: true,
            tag,
            version: parsed.version,
            repo
        }));

    } catch (err) {
        out.error(`Failed to update release: ${err.message}`);
        process.exit(1);
    }
}

main();
