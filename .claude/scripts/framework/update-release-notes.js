#!/usr/bin/env node
// **Version:** 0.20.0
/**
 * @framework-script 0.23.2
 * @description Extract CHANGELOG section and update GitHub Release page
 * @checksum sha256:placeholder
 *
 * This script is provided by the framework and may be updated.
 * To customize: copy to .claude/scripts/shared/ and modify.
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

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

        // Extract section for this version
        const versionPattern = new RegExp(
            `## \\[${version.replace('v', '')}\\][^\\n]*\\n([\\s\\S]*?)(?=## \\[|$)`
        );
        const match = changelog.match(versionPattern);

        if (!match) {
            console.log(JSON.stringify({
                success: false,
                message: `No changelog section found for ${version}`
            }));
            process.exit(1);
        }

        const notes = match[1].trim();

        // Update GitHub release
        const notesFile = path.join(process.cwd(), '.tmp-release-notes.md');
        fs.writeFileSync(notesFile, notes);

        try {
            execSync(`gh release edit ${version} --notes-file "${notesFile}"`, {
                encoding: 'utf8'
            });
        } finally {
            fs.unlinkSync(notesFile);
        }

        console.log(JSON.stringify({
            success: true,
            message: `Updated release notes for ${version}`,
            data: { version }
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
