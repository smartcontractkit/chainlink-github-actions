import * as core from '@actions/core'
import {versionCompare, Operator} from './compare'

/**
 * Simple GitHub Actions wrapper around the semver library.
 */

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
