import * as core from '@actions/core'
import * as fs from 'fs'
import {KebabCasedProperties} from 'type-fest'
import {
  OutputMode,
  TestResult,
  TestResultsSchema,
  TestRunFailures
} from './types'

/**
 * Github action to parse the output of go test -json output to wanted formats with wanted data
 */

async function main(): Promise<void> {
  try {
    const resultsFile: string = getTypedInput('results-file')
    if (!resultsFile) {
      core.error('No results file provided')
    }
    const outputMode: OutputMode = getTypedInput('output-mode') as OutputMode
    const testResults: TestResult[] = deserializeTestResultsFile(resultsFile)
    const failedTests: TestRunFailures = getTestFailures(testResults)
    if (outputMode === 'standard') {
      standardTestOutput(failedTests.TestsFailed, testResults, true)
    } else if (outputMode === 'json') {
      jsonTestOutput(failedTests.TestsFailed, testResults, true)
    }
    if (failedTests.TestsFailed.length > 0) {
      throw new Error('Test Failures Found')
    } else if (
      failedTests.TestsFailed.length === 0 &&
      failedTests.PackageFailure
    ) {
      // we have a failure without a test failure, log out all the output
      logAllOutput(testResults)
      throw new Error('Test Failures Found')
    }
  } catch (error) {
    if (error instanceof Error) core.setFailed(error.message)
  }
}

/**
 * Deserialize the go test results file into a TestResult array
 * @param resultsFile The file with the -json go test results output
 * @returns The TestResult array with all the go test results data
 */
export function deserializeTestResultsFile(resultsFile: string): TestResult[] {
  const fileData = fs.readFileSync(resultsFile, 'utf8')
  // Split the content by newline characters
  const lines = fileData.split('\n').filter(line => line.trim() !== '')

  // Parse each line as a JSON object and collect the parsed objects in an array
  const jsonArray = lines.map(line => JSON.parse(line))
  return TestResultsSchema.parse(jsonArray)
}

/**
 * Get all the failed test names from the test results array
 * @param testResults The test results to filter through
 * @returns The failed test names
 */
export function getTestFailures(testResults: TestResult[]): TestRunFailures {
  const names: string[] = []
  let packageFailure = false
  testResults.filter(testResult => {
    if (testResult.Action === 'fail') {
      if (testResult.Test) {
        names.push(testResult.Test)
      } else {
        // it is possible to have a package failure without a test failure
        // so we are storing whether we have a package failure to determine
        // if we need to output all the output for triage later for panic cases
        packageFailure = true
      }
    }
  })
  return {TestsFailed: names, PackageFailure: packageFailure} as TestRunFailures
}

/**
 * Standard test output.
 * Only failures and just the output in the original format.
 * @param failedTestNames The names of the failed tests to output
 * @param testResults The test results to filter
 * @param shouldLogOutput Should we log the output to the workflow console?
 * @returns The filtered list of test failures outputs
 */
export function standardTestOutput(
  failedTestNames: string[],
  testResults: TestResult[],
  shouldLogOutput: boolean
): TestResult[] {
  let outputString = ''
  const filteredResults: TestResult[] = testResults.filter(testResult => {
    if (testResult.Action === 'output') {
      if (testResult.Test && failedTestNames.includes(testResult.Test)) {
        outputString += testResult.Output
        return testResult
      }
    }
  })
  outputHandler(outputString, shouldLogOutput)
  return filteredResults
}

/**
 * json test output.
 * Only failures and other output that doesn't inclue a test name.
 * Printed in the original broken json array format compatible with gotestfmt
 * @param failedTestNames
 * @param testResults
 * @param shouldLogOutput
 * @returns The filtered list of test result objects
 */
export function jsonTestOutput(
  failedTestNames: string[],
  testResults: TestResult[],
  shouldLogOutput: boolean
): TestResult[] {
  let outputString = ''
  const filteredResults: TestResult[] = testResults.filter(testResult => {
    if (
      testResult.Action === 'output' ||
      testResult.Action === 'run' ||
      testResult.Action === 'fail'
    ) {
      if (testResult.Test) {
        // if we have a test name then we only want tests in the failed test list
        if (failedTestNames.includes(testResult.Test)) {
          outputString += `${JSON.stringify(testResult)}\n`
          return testResult
        }
      } else {
        // if we don't have a test name then we want to just output whatever it is
        outputString += `${JSON.stringify(testResult)}\n`
        return testResult
      }
    } else if (testResult.Action === 'start') {
      // keep this to make gotestfmt happy
      outputString += `${JSON.stringify(testResult)}\n`
      return testResult
    }
  })
  outputHandler(outputString, shouldLogOutput)
  return filteredResults
}

/**
 * Write the output to the workflow console if wanted
 * Also write the output to a file if the filename was provided
 * @param output The string to output
 * @param shouldLog Whether we should log to the console
 */
export function outputHandler(output: string, shouldLog: boolean): void {
  if (shouldLog) {
    core.info(output)
  }
  const outputFile: string = getTypedInput('output-file', false)
  if (outputFile) {
    fs.writeFileSync(outputFile, output)
  }
}

/**
 * We have edge cases where we have a failure but no test failure
 * In this case we want to output all the output for triage since we
 * don't necessarily know what will be useful
 * @param testResults The test results to print outputs from
 */
export function logAllOutput(testResults: TestResult[]): void {
  let outputString = ''
  for (const testResult of testResults) {
    if (testResult.Action === 'output') {
      outputString += testResult.Output
    }
  }
  core.info(
    'We had an error in the test run but no specific test had a failure log, logging out everything for triage'
  )
  core.info(outputString)
}

/**
 * Takes kebob cased inputs and enforces they match the expected inputs
 * @param inputKey The input key to get
 * @param required Is the input required?
 * @returns The input value
 */
function getTypedInput(
  inputKey: keyof KebabCasedProperties<{
    resultsFile: never
    outputMode: never
    outputFile: never
  }>,
  required = true
): string {
  return core.getInput(inputKey, {required, trimWhitespace: true})
}

// Run the main function
main()
