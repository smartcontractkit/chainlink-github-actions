import * as core from '@actions/core'
import semver from 'semver'

/**
 * Simple GitHub Actions wrapper around the semver library.
 */

type Operator = 'gt' | 'lt' | 'eq'

export function versionCompare(
  version1: string,
  operator: Operator,
  version2: string
): boolean {
  // Check for required input:
  if (!version1 || !version2 || !operator) {
    throw new Error('Required inputs not specified.')
  }

  // Validate the inputs:
  if (!semver.valid(version1) || !semver.valid(version2)) {
    throw new Error('Invalid version(s).')
  }

  if (!['gt', 'lt', 'eq'].includes(operator)) {
    throw new Error('Invalid operator.')
  }

  // Compare the versions:
  const result = semver[operator](version1, version2)
  return result
}

async function run(): Promise<void> {
  try {
    const version1 = core.getInput('version1')
    const operator = core.getInput('operator') as Operator
    const version2 = core.getInput('version2')
    core.debug(`Comparing ${version1} ${operator} ${version2}`)

    const result = versionCompare(version1, operator, version2)
    core.setOutput('result', result)
  } catch (error) {
    if (error instanceof Error) core.setFailed(error.message)
  }
}

run()
