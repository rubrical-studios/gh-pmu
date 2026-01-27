/**
 * @framework-script 0.33.2
 * lib/gh.js - GitHub CLI wrapper
 *
 * Provides functions for interacting with GitHub via the gh CLI.
 */

const { execSync, spawnSync } = require('child_process');

/**
 * Check if an error message indicates a transient HTTP error (5xx)
 * @param {string} message - Error message to check
 * @returns {boolean} True if transient error
 */
function isTransientError(message) {
    // Match HTTP 5xx errors (502, 503, 504, etc.)
    return /HTTP 5\d{2}/.test(message);
}

/**
 * Synchronous sleep using spawnSync
 * @param {number} ms - Milliseconds to sleep
 */
function sleepSync(ms) {
    if (process.platform === 'win32') {
        spawnSync('ping', ['-n', Math.ceil(ms / 1000) + 1, '127.0.0.1'], { stdio: 'ignore' });
    } else {
        spawnSync('sleep', [ms / 1000], { stdio: 'ignore' });
    }
}

/**
 * Execute a gh command and return parsed JSON
 * @param {string} args - gh command arguments
 * @param {object} options - Execution options
 * @param {number} options.retries - Max retries for transient errors (default: 3)
 * @param {number} options.retryDelay - Initial delay between retries in ms (default: 1000)
 * @returns {object} Parsed JSON response
 * @throws {Error} If gh command fails after all retries
 */
function exec(args, options = {}) {
    const { retries = 3, retryDelay = 1000 } = options;
    let lastError;
    let delay = retryDelay;

    for (let attempt = 0; attempt <= retries; attempt++) {
        try {
            const output = execSync(`gh ${args}`, { encoding: 'utf8' });
            return output.trim();
        } catch (err) {
            lastError = err;

            // Only retry on transient errors
            if (attempt < retries && isTransientError(err.message)) {
                sleepSync(delay);
                delay *= 2; // Exponential backoff
                continue;
            }

            throw new Error(`gh ${args} failed: ${err.message}`);
        }
    }

    throw new Error(`gh ${args} failed after ${retries} retries: ${lastError.message}`);
}

/**
 * Execute a gh command and return parsed JSON
 * @param {string} args - gh command arguments
 * @returns {object} Parsed JSON response
 */
function execJson(args) {
    const output = exec(args);
    try {
        return JSON.parse(output);
    } catch {
        throw new Error(`Failed to parse gh output as JSON: ${output}`);
    }
}

/**
 * Get the latest workflow run
 * @param {string} repo - Repository in owner/repo format (optional)
 * @returns {object} Latest run info
 */
function getLatestRun(repo) {
    const repoArg = repo ? `--repo ${repo}` : '';
    const runs = execJson(`run list --limit 1 --json databaseId,status,conclusion,name,headBranch,createdAt,updatedAt ${repoArg}`);
    return runs[0] || null;
}

/**
 * Get workflow runs with optional filtering
 * @param {object} options - Filter options
 * @param {number} options.limit - Max runs to return
 * @param {string} options.branch - Filter by branch
 * @param {string} options.repo - Repository
 * @returns {Array} Array of runs
 */
function getRuns(options = {}) {
    const { limit = 10, branch, repo } = options;
    let args = `run list --limit ${limit} --json databaseId,status,conclusion,name,headBranch,createdAt,updatedAt`;
    if (branch) args += ` --branch ${branch}`;
    if (repo) args += ` --repo ${repo}`;
    return execJson(args);
}

/**
 * Get a specific workflow run by ID
 * @param {number|string} runId - The run ID
 * @param {string} repo - Repository in owner/repo format (optional)
 * @returns {object} Run details with jobs
 */
function getRun(runId, repo) {
    const repoArg = repo ? `--repo ${repo}` : '';
    return execJson(`run view ${runId} --json databaseId,status,conclusion,name,jobs,createdAt,updatedAt ${repoArg}`);
}

/**
 * Get release information by tag
 * @param {string} tag - The release tag
 * @param {string} repo - Repository in owner/repo format (optional)
 * @returns {object} Release info with assets
 */
function getRelease(tag, repo) {
    const repoArg = repo ? `--repo ${repo}` : '';
    try {
        return execJson(`release view ${tag} --json tagName,name,body,createdAt,assets,url ${repoArg}`);
    } catch {
        // Release might not exist yet
        return null;
    }
}

/**
 * List releases
 * @param {object} options - Options
 * @param {number} options.limit - Max releases to return
 * @param {string} options.repo - Repository
 * @returns {Array} Array of releases
 */
function listReleases(options = {}) {
    const { limit = 10, repo } = options;
    const repoArg = repo ? `--repo ${repo}` : '';
    try {
        return execJson(`release list --limit ${limit} --json tagName,name,createdAt,isDraft,isPrerelease ${repoArg}`);
    } catch {
        return [];
    }
}

/**
 * Get the current repository in owner/repo format
 * @returns {string} Repository identifier
 */
function getCurrentRepo() {
    try {
        const output = exec('repo view --json nameWithOwner -q .nameWithOwner');
        return output;
    } catch {
        return null;
    }
}

/**
 * Check if gh CLI is available and authenticated
 * @returns {boolean}
 */
function isAvailable() {
    try {
        exec('auth status');
        return true;
    } catch {
        return false;
    }
}

/**
 * Get workflow run triggered by a tag push
 * @param {string} tag - The tag name
 * @param {string} repo - Repository (optional)
 * @returns {object|null} The run or null if not found
 */
function getTagTriggeredRun(tag, repo) {
    const runs = getRuns({ limit: 20, repo });

    // Look for a run on the tag ref
    for (const run of runs) {
        if (run.headBranch === tag || run.headBranch === `refs/tags/${tag}`) {
            return run;
        }
    }

    return null;
}

module.exports = {
    exec,
    execJson,
    getLatestRun,
    getRuns,
    getRun,
    getRelease,
    listReleases,
    getCurrentRepo,
    isAvailable,
    getTagTriggeredRun
};
