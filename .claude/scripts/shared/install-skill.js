#!/usr/bin/env node
/**
 * @framework-script 0.42.0
 * install-skill.js - Skill package extraction and deployment
 *
 * Shared script for deploying skills from framework packages to projects.
 * Used by /charter skill selection and /install-skill command.
 *
 * Usage:
 *   node install-skill.js <skill-name> [--framework-path <path>] [--project-dir <path>]
 *   node install-skill.js --list [--installed]
 *   node install-skill.js --help
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

// ============================================================================
// ZIP EXTRACTION (reused from install/lib/detection.js)
// ============================================================================

function extractZip(zipPath, destDir) {
  try {
    fs.mkdirSync(destDir, { recursive: true });

    if (process.platform === 'win32') {
      // PowerShell Expand-Archive
      execSync(`powershell -Command "Expand-Archive -Path '${zipPath}' -DestinationPath '${destDir}' -Force"`, { stdio: 'pipe' });
    } else {
      // Unix unzip
      execSync(`unzip -q -o "${zipPath}" -d "${destDir}"`, { stdio: 'pipe' });
    }
    return true;
  } catch (_err) {
    return false;
  }
}

// ============================================================================
// SKILL INSTALLATION
// ============================================================================

/**
 * Install a skill from the framework to a project.
 * @param {string} frameworkPath - Path to the framework directory
 * @param {string} projectDir - Path to the project directory
 * @param {string} skillName - Name of the skill to install
 * @returns {object} { status: 'installed'|'skipped'|'failed', reason?: string, resources?: number }
 */
function installSkill(frameworkPath, projectDir, skillName) {
  const packagePath = path.join(frameworkPath, 'Skills', 'Packaged', `${skillName}.zip`);
  const destPath = path.join(projectDir, '.claude', 'skills', skillName);
  const skillMdPath = path.join(destPath, 'SKILL.md');

  // Check if already installed
  if (fs.existsSync(skillMdPath)) {
    return { status: 'skipped', reason: 'already installed' };
  }

  // Check package exists
  if (!fs.existsSync(packagePath)) {
    return { status: 'failed', reason: 'package not found' };
  }

  // Ensure .claude/skills directory exists
  const skillsDir = path.join(projectDir, '.claude', 'skills');
  if (!fs.existsSync(skillsDir)) {
    fs.mkdirSync(skillsDir, { recursive: true });
  }

  // Extract
  if (!extractZip(packagePath, destPath)) {
    return { status: 'failed', reason: 'extraction failed' };
  }

  // Verify SKILL.md exists after extraction
  if (!fs.existsSync(skillMdPath)) {
    return { status: 'failed', reason: 'SKILL.md not found after extraction' };
  }

  // Count resources (files in the skill directory)
  const resources = countResources(destPath);

  return { status: 'installed', resources };
}

/**
 * Count files in a directory (excluding SKILL.md itself).
 */
function countResources(dir) {
  let count = 0;
  try {
    const entries = fs.readdirSync(dir, { withFileTypes: true });
    for (const entry of entries) {
      if (entry.isFile() && entry.name !== 'SKILL.md') {
        count++;
      } else if (entry.isDirectory()) {
        count += countResources(path.join(dir, entry.name));
      }
    }
  } catch (_err) {
    // Directory doesn't exist or not readable
  }
  return count;
}

/**
 * Add skills to projectSkills in framework-config.json.
 * @param {string} projectDir - Path to the project directory
 * @param {string[]} skillNames - Array of skill names to add
 */
function addToProjectSkills(projectDir, skillNames) {
  const configPath = path.join(projectDir, 'framework-config.json');
  let config = {};

  if (fs.existsSync(configPath)) {
    try {
      config = JSON.parse(fs.readFileSync(configPath, 'utf-8'));
    } catch (_err) {
      // Invalid JSON, start fresh
    }
  }

  // Initialize projectSkills if not present
  if (!Array.isArray(config.projectSkills)) {
    config.projectSkills = [];
  }

  // Add new skills (avoid duplicates)
  for (const skill of skillNames) {
    if (!config.projectSkills.includes(skill)) {
      config.projectSkills.push(skill);
    }
  }

  // Sort for consistency
  config.projectSkills.sort();

  // Write back
  fs.writeFileSync(configPath, JSON.stringify(config, null, 2) + '\n');
}

/**
 * Get framework path from project's framework-config.json.
 * @param {string} projectDir - Path to the project directory
 * @returns {string|null} Framework path or null if not found
 */
function getFrameworkPath(projectDir) {
  const configPath = path.join(projectDir, 'framework-config.json');

  if (!fs.existsSync(configPath)) {
    return null;
  }

  try {
    const config = JSON.parse(fs.readFileSync(configPath, 'utf-8'));
    return config.frameworkPath || null;
  } catch (_err) {
    return null;
  }
}

/**
 * List available skills from skill-registry.json.
 * @param {string} projectDir - Path to the project directory
 * @returns {object[]} Array of { name, description, installed }
 */
function listSkills(projectDir) {
  const registryPath = path.join(projectDir, '.claude', 'metadata', 'skill-registry.json');
  const skillsDir = path.join(projectDir, '.claude', 'skills');

  if (!fs.existsSync(registryPath)) {
    return [];
  }

  try {
    const registry = JSON.parse(fs.readFileSync(registryPath, 'utf-8'));
    return registry.skills.map(skill => ({
      name: skill.name,
      description: skill.description,
      installed: fs.existsSync(path.join(skillsDir, skill.name, 'SKILL.md'))
    }));
  } catch (_err) {
    return [];
  }
}

/**
 * List installed skills by scanning .claude/skills directory.
 * @param {string} projectDir - Path to the project directory
 * @returns {string[]} Array of installed skill names
 */
function listInstalledSkills(projectDir) {
  const skillsDir = path.join(projectDir, '.claude', 'skills');

  if (!fs.existsSync(skillsDir)) {
    return [];
  }

  const installed = [];
  try {
    const entries = fs.readdirSync(skillsDir, { withFileTypes: true });
    for (const entry of entries) {
      if (entry.isDirectory()) {
        const skillMd = path.join(skillsDir, entry.name, 'SKILL.md');
        if (fs.existsSync(skillMd)) {
          installed.push(entry.name);
        }
      }
    }
  } catch (_err) {
    // Directory not readable
  }

  return installed.sort();
}

// ============================================================================
// CLI
// ============================================================================

function showHelp() {
  console.log(`
Usage: node install-skill.js <skill-name> [options]
       node install-skill.js --list [--installed]

Options:
  --framework-path <path>  Path to framework directory (default: from framework-config.json)
  --project-dir <path>     Path to project directory (default: current directory)
  --list                   List available skills
  --installed              With --list, show installation status
  --help                   Show this help message

Examples:
  node install-skill.js electron-development
  node install-skill.js playwright-setup postgresql-integration
  node install-skill.js --list
  node install-skill.js --list --installed
`);
}

function main() {
  const args = process.argv.slice(2);

  if (args.includes('--help') || args.length === 0) {
    showHelp();
    process.exit(0);
  }

  // Parse options
  let projectDir = process.cwd();
  let frameworkPath = null;
  let listMode = false;
  let showInstalled = false;
  const skillNames = [];

  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (arg === '--project-dir' && args[i + 1]) {
      projectDir = args[++i];
    } else if (arg === '--framework-path' && args[i + 1]) {
      frameworkPath = args[++i];
    } else if (arg === '--list') {
      listMode = true;
    } else if (arg === '--installed') {
      showInstalled = true;
    } else if (!arg.startsWith('-')) {
      skillNames.push(arg);
    }
  }

  // Get framework path if not specified
  if (!frameworkPath) {
    frameworkPath = getFrameworkPath(projectDir);
  }

  // List mode
  if (listMode) {
    const skills = listSkills(projectDir);

    if (skills.length === 0) {
      console.log('No skill registry found. Run the framework installer first.');
      process.exit(1);
    }

    console.log('\nAvailable Skills:\n');
    console.log('Name                          Description                                          Status');
    console.log('─'.repeat(100));

    for (const skill of skills) {
      const name = skill.name.padEnd(30);
      const desc = (skill.description || '').substring(0, 50).padEnd(52);
      const status = showInstalled ? (skill.installed ? 'Installed' : 'Available') : '';
      console.log(`${name}${desc}${status}`);
    }

    console.log(`\nTotal: ${skills.length} skills`);
    if (showInstalled) {
      const installedCount = skills.filter(s => s.installed).length;
      console.log(`Installed: ${installedCount}`);
    }

    process.exit(0);
  }

  // Install mode
  if (skillNames.length === 0) {
    console.error('Error: No skill names provided.');
    showHelp();
    process.exit(1);
  }

  if (!frameworkPath) {
    console.error('Error: Cannot determine framework path. Specify --framework-path or ensure framework-config.json exists.');
    process.exit(1);
  }

  console.log('\nInstalling skills...\n');

  let installed = 0;
  let skipped = 0;
  let failed = 0;
  const installedSkills = [];

  for (const skillName of skillNames) {
    const result = installSkill(frameworkPath, projectDir, skillName);

    if (result.status === 'installed') {
      console.log(`\u2713 ${skillName} - Installed (${result.resources} resources)`);
      installed++;
      installedSkills.push(skillName);
    } else if (result.status === 'skipped') {
      console.log(`\u2298 ${skillName} - Already installed (skipped)`);
      skipped++;
    } else {
      console.log(`\u2717 ${skillName} - ${result.reason}`);
      failed++;
    }
  }

  // Add installed skills to framework-config.json
  if (installedSkills.length > 0) {
    addToProjectSkills(projectDir, installedSkills);
  }

  console.log(`\nInstalled: ${installed}  Skipped: ${skipped}  Failed: ${failed}`);

  process.exit(failed > 0 ? 1 : 0);
}

// Export functions for use as module
module.exports = {
  installSkill,
  addToProjectSkills,
  getFrameworkPath,
  listSkills,
  listInstalledSkills,
  extractZip,
};

// Run CLI if executed directly
if (require.main === module) {
  main();
}
