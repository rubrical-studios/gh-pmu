#!/usr/bin/env node
/**
 * @framework-script 0.42.0
 * cleanup-release-assets.js - Clean up old release assets
 *
 * Deletes binary assets from older GitHub releases, keeping only the
 * N most recent tagged releases (default: 3).
 *
 * Usage:
 *   node cleanup-release-assets.js
 *   node cleanup-release-assets.js --keep 5
 *   node cleanup-release-assets.js --dry-run
 *
 * Options:
 *   --keep <n>            Number of recent releases to keep assets for (default: 3)
 *   --dry-run             Show what would be deleted without making changes
 *   --quiet               Suppress non-output messages
 *   --help                Show this help message
 *
 * Exit codes:
 *   0 - Success
 *   1 - Error
 */

const gh = require('./lib/gh');
const out = require('./lib/output');

function showHelp() {
    console.log(`
cleanup-release-assets.js - Clean up old release assets

Usage:
  node cleanup-release-assets.js
  node cleanup-release-assets.js --keep 5
  node cleanup-release-assets.js --dry-run

Options:
  --keep <n>            Number of recent releases to keep assets for (default: 3)
  --dry-run             Show what would be deleted without making changes
  --quiet               Suppress non-output messages
  --help                Show this help message

This script:
1. Lists all GitHub releases sorted by date (newest first)
2. Keeps assets for the N most recent tagged releases
3. Deletes assets from older releases (preserves the release entries)

Example:
  Releases: v0.9.1, v0.9.0, v0.8.6, v0.8.5, v0.8.4, v0.8.3, ...
  With --keep 3: Keeps assets for v0.9.1, v0.9.0, v0.8.6
                 Deletes assets from v0.8.5 and older
`);
}

/**
 * Get all releases with their assets
 * @param {string} repo - Repository in owner/repo format
 * @returns {Array} Array of releases with assets
 */
function getAllReleases(repo) {
    try {
        // Get more releases to ensure we capture all
        const releases = gh.execJson(`release list --limit 100 --json tagName,name,createdAt,isDraft,isPrerelease`);

        // Filter out drafts and get asset details for each
        const releasesWithAssets = [];

        for (const release of releases) {
            if (release.isDraft) continue;

            try {
                const details = gh.getRelease(release.tagName, repo);
                if (details && details.assets && details.assets.length > 0) {
                    releasesWithAssets.push({
                        tagName: release.tagName,
                        name: release.name || release.tagName,
                        createdAt: release.createdAt,
                        isPrerelease: release.isPrerelease,
                        assets: details.assets
                    });
                }
            } catch {
                // Skip releases that fail to fetch
            }
        }

        return releasesWithAssets;
    } catch (err) {
        throw new Error(`Failed to list releases: ${err.message}`);
    }
}

/**
 * Delete an asset from a release
 * @param {string} tagName - Release tag
 * @param {string} assetName - Asset filename
 * @param {boolean} dryRun - If true, don't actually delete
 * @returns {boolean} Success status
 */
function deleteAsset(tagName, assetName, dryRun = false) {
    if (dryRun) {
        return true;
    }

    try {
        gh.exec(`release delete-asset ${tagName} "${assetName}" --yes`);
        return true;
    } catch {
        return false;
    }
}

async function main() {
    const flags = out.parseFlags();

    if (flags.help) {
        showHelp();
        process.exit(0);
    }

    const keepCount = parseInt(out.getFlag(flags.args, '--keep') || '3', 10);
    const dryRun = flags.args.includes('--dry-run');
    const quiet = flags.args.includes('--quiet');

    if (isNaN(keepCount) || keepCount < 1) {
        out.error('--keep must be a positive number');
        process.exit(1);
    }

    // Get repository info
    const repo = gh.getCurrentRepo();
    if (!repo) {
        out.error('Could not determine repository. Run from a git repository.');
        process.exit(1);
    }

    if (!quiet) {
        out.info(`Analyzing releases for ${repo}...`);
        out.info(`Will keep assets for ${keepCount} most recent releases`);
        if (dryRun) {
            out.info('Dry run mode - no changes will be made');
        }
    }

    // Get all releases with assets
    const releases = getAllReleases(repo);

    if (releases.length === 0) {
        if (!quiet) {
            out.info('No releases with assets found');
        }
        process.exit(0);
    }

    // Sort by date (newest first)
    releases.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));

    if (!quiet) {
        out.info(`Found ${releases.length} releases with assets`);
    }

    // Separate releases to keep and cleanup
    const toKeep = releases.slice(0, keepCount);
    const toCleanup = releases.slice(keepCount);

    if (!quiet && toKeep.length > 0) {
        console.log('\nKeeping assets for:');
        for (const release of toKeep) {
            console.log(`  - ${release.tagName} (${release.assets.length} assets)`);
        }
    }

    if (toCleanup.length === 0) {
        if (!quiet) {
            out.success('No old release assets to clean up');
        }
        console.log(JSON.stringify({
            success: true,
            kept: toKeep.map(r => r.tagName),
            cleaned: [],
            assetsDeleted: 0
        }));
        process.exit(0);
    }

    if (!quiet) {
        console.log('\nCleaning up assets from:');
    }

    let totalDeleted = 0;
    const cleanedReleases = [];
    const errors = [];

    for (const release of toCleanup) {
        if (!quiet) {
            console.log(`  - ${release.tagName} (${release.assets.length} assets)`);
        }

        let releaseDeleted = 0;
        for (const asset of release.assets) {
            const success = deleteAsset(release.tagName, asset.name, dryRun);
            if (success) {
                releaseDeleted++;
                totalDeleted++;
                if (!quiet && !dryRun) {
                    console.log(`    Deleted: ${asset.name}`);
                } else if (!quiet && dryRun) {
                    console.log(`    Would delete: ${asset.name}`);
                }
            } else {
                errors.push(`${release.tagName}/${asset.name}`);
            }
        }

        cleanedReleases.push({
            tagName: release.tagName,
            assetsDeleted: releaseDeleted
        });
    }

    // Summary
    if (!quiet) {
        console.log('');
        if (dryRun) {
            out.info(`Would delete ${totalDeleted} assets from ${toCleanup.length} releases`);
        } else {
            out.success(`Deleted ${totalDeleted} assets from ${toCleanup.length} releases`);
        }

        if (errors.length > 0) {
            out.warn(`Failed to delete ${errors.length} assets`);
        }
    }

    // Output JSON for programmatic use
    console.log(JSON.stringify({
        success: true,
        dryRun,
        kept: toKeep.map(r => r.tagName),
        cleaned: cleanedReleases,
        assetsDeleted: totalDeleted,
        errors: errors.length > 0 ? errors : undefined
    }));
}

main();
