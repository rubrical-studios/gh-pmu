#!/usr/bin/env node
/**
 * @framework-script 0.33.3
 * @description Extract CHANGELOG section and update GitHub Release page with formatted notes
 * @checksum sha256:placeholder
 *
 * This script transforms raw CHANGELOG entries into a standardized release page format:
 * - Title: "IDPF Framework vX.Y.Z"
 * - Release date
 * - Auto-generated summary
 * - Category sections (Added, Fixed, Changed, etc.)
 * - Full changelog comparison link
 *
 * This script is provided by the framework and may be updated.
 * To customize: copy to .claude/scripts/shared/ and modify.
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

/**
 * Check if a GitHub release exists for the given version
 * @param {string} version - The version tag (e.g., v1.0.0)
 * @returns {boolean} - True if release exists, false otherwise
 */
function releaseExists(version) {
    try {
        execSync(`gh release view ${version} --json tagName`, {
            encoding: 'utf8',
            stdio: ['pipe', 'pipe', 'pipe']
        });
        return true;
    } catch {
        return false;
    }
}

/**
 * Sleep for specified milliseconds
 * @param {number} ms - Milliseconds to sleep
 */
function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Update or create a GitHub release with retry logic
 * Handles race conditions where GitHub Actions may create the release
 * @param {string} version - The version tag
 * @param {string} notesFile - Path to the release notes file
 * @param {number} maxRetries - Maximum number of retries (default: 3)
 * @param {number} retryDelayMs - Delay between retries in ms (default: 2000)
 */
async function updateOrCreateRelease(version, notesFile, maxRetries = 3, retryDelayMs = 2000) {
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
        try {
            if (releaseExists(version)) {
                // Release exists - update it
                execSync(`gh release edit ${version} --notes-file "${notesFile}"`, {
                    encoding: 'utf8',
                    stdio: ['pipe', 'pipe', 'pipe']
                });
                return { action: 'updated' };
            } else {
                // Release doesn't exist - create it
                execSync(`gh release create ${version} --title "Release ${version}" --notes-file "${notesFile}"`, {
                    encoding: 'utf8',
                    stdio: ['pipe', 'pipe', 'pipe']
                });
                return { action: 'created' };
            }
        } catch (err) {
            const errorMsg = err.message || '';

            // If "already exists" error, the release was created between our check and create
            // Retry with edit instead
            if (errorMsg.includes('already exists') && attempt < maxRetries) {
                await sleep(retryDelayMs);
                continue;
            }

            // If "release not found" on edit, it may not have propagated yet
            // Retry after delay
            if (errorMsg.includes('release not found') && attempt < maxRetries) {
                await sleep(retryDelayMs);
                continue;
            }

            // Final attempt or unrecoverable error
            if (attempt === maxRetries) {
                throw new Error(`Failed after ${maxRetries} attempts: ${err.message}`);
            }
        }
    }
}

/**
 * Get repository URL from git remote
 */
function getRepoUrl() {
    try {
        const remote = execSync('git remote get-url origin', { encoding: 'utf8' }).trim();
        // Convert SSH to HTTPS format if needed
        if (remote.startsWith('git@github.com:')) {
            return remote.replace('git@github.com:', 'https://github.com/').replace(/\.git$/, '');
        }
        return remote.replace(/\.git$/, '');
    } catch {
        return null;
    }
}

/**
 * Get previous tag before the current one
 */
function getPreviousTag(currentTag) {
    try {
        // Get all tags sorted by version, find the one before current
        const tags = execSync('git tag --sort=-v:refname', { encoding: 'utf8' })
            .trim()
            .split('\n')
            .filter(t => t.match(/^v\d+\.\d+\.\d+$/));

        const currentIndex = tags.indexOf(currentTag);
        if (currentIndex >= 0 && currentIndex < tags.length - 1) {
            return tags[currentIndex + 1];
        }
        return null;
    } catch {
        return null;
    }
}

/**
 * Count items in each category for summary generation
 */
function countCategoryItems(content) {
    const counts = {};
    const categoryRegex = /^### (\w+)/gm;
    let match;

    while ((match = categoryRegex.exec(content)) !== null) {
        const category = match[1];
        // Find the section content and count bullet points
        const sectionStart = match.index + match[0].length;
        const nextSection = content.slice(sectionStart).search(/^### \w+/m);
        const sectionContent = nextSection >= 0
            ? content.slice(sectionStart, sectionStart + nextSection)
            : content.slice(sectionStart);

        // Count top-level bullet points (lines starting with "- ")
        const items = sectionContent.match(/^- /gm) || [];
        counts[category] = items.length;
    }

    return counts;
}

/**
 * Generate summary based on category counts
 */
function generateSummary(counts) {
    const parts = [];

    if (counts.Added) {
        parts.push(`${counts.Added} new feature${counts.Added > 1 ? 's' : ''}`);
    }
    if (counts.Changed) {
        parts.push(`${counts.Changed} change${counts.Changed > 1 ? 's' : ''}`);
    }
    if (counts.Fixed) {
        parts.push(`${counts.Fixed} fix${counts.Fixed > 1 ? 'es' : ''}`);
    }
    if (counts.Removed) {
        parts.push(`${counts.Removed} removal${counts.Removed > 1 ? 's' : ''}`);
    }
    if (counts.Security) {
        parts.push(`${counts.Security} security update${counts.Security > 1 ? 's' : ''}`);
    }
    if (counts.Deprecated) {
        parts.push(`${counts.Deprecated} deprecation${counts.Deprecated > 1 ? 's' : ''}`);
    }
    if (counts.Documentation) {
        parts.push(`documentation updates`);
    }

    if (parts.length === 0) {
        return 'Maintenance release.';
    }

    // Determine release type
    const hasFeatures = counts.Added > 0;
    const hasOnlyFixes = !hasFeatures && counts.Fixed > 0;

    let prefix = '';
    if (hasOnlyFixes && Object.keys(counts).length <= 2) {
        prefix = 'Patch release with ';
    } else if (hasFeatures) {
        prefix = 'Feature release with ';
    } else {
        prefix = 'Release with ';
    }

    // Join parts with proper grammar
    if (parts.length === 1) {
        return prefix + parts[0] + '.';
    } else if (parts.length === 2) {
        return prefix + parts.join(' and ') + '.';
    } else {
        const last = parts.pop();
        return prefix + parts.join(', ') + ', and ' + last + '.';
    }
}

/**
 * Transform CHANGELOG content to release page format
 */
function transformToReleaseFormat(version, date, rawContent, repoUrl, previousTag) {
    // Promote heading levels: ### Category -> ## Category
    let content = rawContent.replace(/^### /gm, '## ');

    // Remove trailing horizontal rules
    content = content.replace(/\n---\s*$/, '');

    // Generate summary
    const counts = countCategoryItems(rawContent);
    const summary = generateSummary(counts);

    // Build formatted release notes
    let notes = `# IDPF Framework ${version}\n\n`;
    notes += `**Release Date:** ${date}\n\n`;
    notes += `## Summary\n\n${summary}\n\n`;
    notes += content.trim();

    // Add comparison link if we have repo URL and previous tag
    if (repoUrl && previousTag) {
        notes += `\n\n---\n\n`;
        notes += `**Full Changelog:** [${previousTag}...${version}](${repoUrl}/compare/${previousTag}...${version})`;
    }

    return notes;
}

async function main() {
    // Get version from args or try to detect from latest tag
    let version = process.argv[2];

    if (!version) {
        try {
            version = execSync('git describe --tags --abbrev=0', {
                encoding: 'utf8'
            }).trim();
        } catch {
            console.log(JSON.stringify({
                success: false,
                message: 'Version not provided and no tags found'
            }));
            process.exit(1);
        }
    }

    // Ensure version has 'v' prefix for consistency
    if (!version.startsWith('v')) {
        version = 'v' + version;
    }

    try {
        // Read CHANGELOG.md
        const changelogPath = path.join(process.cwd(), 'CHANGELOG.md');
        if (!fs.existsSync(changelogPath)) {
            console.log(JSON.stringify({
                success: false,
                message: 'CHANGELOG.md not found'
            }));
            process.exit(1);
        }

        const changelog = fs.readFileSync(changelogPath, 'utf8');
        const versionNum = version.replace('v', '');

        // Extract section for this version including the header (for date)
        const versionPattern = new RegExp(
            `## \\[${versionNum}\\]\\s*-\\s*(\\d{4}-\\d{2}-\\d{2})\\n([\\s\\S]*?)(?=## \\[|$)`
        );
        const match = changelog.match(versionPattern);

        if (!match) {
            console.log(JSON.stringify({
                success: false,
                message: `No changelog section found for ${version}`
            }));
            process.exit(1);
        }

        const date = match[1];
        const rawContent = match[2].trim();

        // Get repo URL and previous tag for comparison link
        const repoUrl = getRepoUrl();
        const previousTag = getPreviousTag(version);

        // Transform to release format
        const notes = transformToReleaseFormat(version, date, rawContent, repoUrl, previousTag);

        // Update GitHub release
        const notesFile = path.join(process.cwd(), '.tmp-release-notes.md');
        fs.writeFileSync(notesFile, notes);

        let result;
        try {
            result = await updateOrCreateRelease(version, notesFile);
        } finally {
            fs.unlinkSync(notesFile);
        }

        console.log(JSON.stringify({
            success: true,
            message: `${result.action === 'created' ? 'Created' : 'Updated'} release notes for ${version}`,
            data: {
                version,
                date,
                previousTag,
                action: result.action,
                summary: generateSummary(countCategoryItems(rawContent))
            }
        }));

    } catch (err) {
        console.log(JSON.stringify({
            success: false,
            message: `Failed to update release notes: ${err.message}`
        }));
        process.exit(1);
    }
}

main();
