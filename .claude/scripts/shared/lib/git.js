// **Version:** 0.29.3
/**
 * lib/git.js - Git command utilities
 *
 * Provides functions for interacting with git repositories.
 */

const { execSync } = require('child_process');

/**
 * Execute a git command and return the output
 * @param {string} args - Git command arguments
 * @returns {string} Command output
 * @throws {Error} If git command fails
 */
function exec(args) {
    try {
        return execSync(`git ${args}`, { encoding: 'utf8' }).trim();
    } catch (err) {
        throw new Error(`git ${args} failed: ${err.message}`);
    }
}

/**
 * Get the latest semver tag
 * @returns {string|null} Latest tag or null if no tags exist
 */
function getLatestTag() {
    try {
        const tag = exec('tag --sort=-v:refname');
        const tags = tag.split('\n').filter(t => t.trim());
        // Find first tag that looks like a version (v1.2.3 or 1.2.3)
        for (const t of tags) {
            if (/^v?\d+\.\d+\.\d+/.test(t)) {
                return t;
            }
        }
        return tags[0] || null;
    } catch {
        return null;
    }
}

/**
 * Get commits since a specific tag
 * @param {string} tag - The tag to start from
 * @returns {Array<{hash: string, message: string}>} Array of commits
 */
function getCommitsSince(tag) {
    try {
        const range = tag ? `${tag}..HEAD` : 'HEAD';
        const output = exec(`log ${range} --format="%H|%s"`);
        if (!output) return [];

        return output.split('\n')
            .filter(line => line.trim())
            .map(line => {
                const [hash, ...messageParts] = line.split('|');
                return {
                    hash: hash.substring(0, 7),
                    message: messageParts.join('|')
                };
            });
    } catch {
        return [];
    }
}

/**
 * Parse a conventional commit message
 * @param {string} message - The commit message
 * @returns {{type: string, scope: string|null, message: string, breaking: boolean}}
 */
function parseConventionalCommit(message) {
    // Match: type(scope)!: message or type!: message or type(scope): message or type: message
    const match = message.match(/^(\w+)(?:\(([^)]+)\))?(!)?\s*:\s*(.+)$/);

    if (!match) {
        return {
            type: 'other',
            scope: null,
            message: message,
            breaking: false
        };
    }

    const [, type, scope, bang, msg] = match;
    const breaking = !!bang || message.includes('BREAKING CHANGE');

    return {
        type: type.toLowerCase(),
        scope: scope || null,
        message: msg,
        breaking
    };
}

/**
 * Check if a file has uncommitted changes
 * @param {string} file - Path to the file
 * @returns {boolean} True if file is dirty
 */
function isDirty(file) {
    try {
        const status = exec(`status --porcelain "${file}"`);
        return status.length > 0;
    } catch {
        return false;
    }
}

/**
 * Get the diff for a file
 * @param {string} file - Path to the file
 * @returns {string} Diff output
 */
function getDiff(file) {
    try {
        return exec(`diff "${file}"`);
    } catch {
        return '';
    }
}

/**
 * Checkout a file to restore it to committed state
 * @param {string} file - Path to the file
 */
function checkout(file) {
    exec(`checkout "${file}"`);
}

/**
 * Get the current branch name
 * @returns {string} Branch name
 */
function getCurrentBranch() {
    return exec('rev-parse --abbrev-ref HEAD');
}

/**
 * Check if we're in a git repository
 * @returns {boolean}
 */
function isGitRepo() {
    try {
        exec('rev-parse --git-dir');
        return true;
    } catch {
        return false;
    }
}

module.exports = {
    getLatestTag,
    getCommitsSince,
    parseConventionalCommit,
    isDirty,
    getDiff,
    checkout,
    getCurrentBranch,
    isGitRepo
};
