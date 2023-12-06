import * as core from '@actions/core'
import * as fs from 'fs'
import * as readline from 'readline'
import {KebabCasedProperties} from 'type-fest'
import {PassFailMapItem, TestResult, TestResultSchema} from './types'

/**
 * Github action to parse the output of go test -json output to show only failure information
 */

async function main(): Promise<void> {
  try {
    const resultsFile: string = getTypedInput('results-file')
    if (!resultsFile) {
      core.error('No results file provided')
    }
    const outputFile: string = getTypedInput('output-file', false)

    // We have output files that are too large to read in to memory so we need to stream them
    // Go test information can be in strange orders at times, especially when panics happen
    // We need to process what packages and tests actually failed in 1 pass and then write out
    // the wanted output in a second pass

    // read through the file once to preprocess what tests and packages actually need written out
    const failureMap: Map<string, PassFailMapItem> =
      await preprocessTestResultsFile(resultsFile)

    // read through a second time and write out the wanted output
    const passed: boolean = await processTestResultsFile(
      failureMap,
      resultsFile,
      true,
      outputFile
    )

    // throw an error if we have any failures so the action fails
    if (!passed) {
      throw new Error('Test Failures Found')
    }
  } catch (error) {
    if (error instanceof Error) core.setFailed(error.message)
  }
}

/**
 * Read through the file as a stream and build a map of package and test failures
 * @param resultsFile The file to parse
 * @returns The package map with failure information needed to write out the wanted output
 */
export async function preprocessTestResultsFile(
  resultsFile: string
): Promise<Map<string, PassFailMapItem>> {
  // create a map to hold packages that failed and tests that failed for that package
  const failureMap = new Map<string, PassFailMapItem>()
  // setup a stream to read the file in line by line
  const stream1 = fs.createReadStream(resultsFile)
  const reader1 = readline.createInterface({input: stream1})

  for await (const line of reader1) {
    let result: TestResult
    try {
      const parsed = JSON.parse(line)
      result = TestResultSchema.parse(parsed)
    } catch (error) {
      // we ran into errors parsing the json, ignore in the preprocessing step
      continue
    }

    // if we don't have package information then we can skip it
    if (!result.Package) {
      continue
    }

    // we have a package, see if it is in the map and if not add it
    const key = result.Package
    if (!failureMap.has(key)) {
      failureMap.set(key, {
        tests: [],
        packageFailButNoTestFail: false,
        panicTestsFound: []
      })
    }

    // get the item from the map to use to update with data below
    const item = failureMap.get(key)
    if (!item) {
      continue
    }

    // if we have a test then we need to add it to the maps array at that key
    if (result.Test) {
      // do we already have the test in the array
      let foundIndex = -1
      for (let i = 0; i < item.tests.length; i++) {
        if (item.tests[i].test === result.Test) {
          foundIndex = i
        }
      }

      // if not then add a failing test in to the array
      if (foundIndex === -1) {
        foundIndex = item.tests.length
        item.tests.push({test: result.Test, pass: false, completed: false})
      }

      // if the test passed or failed we mark it as so and also mark the test as completed
      if (result.Action === 'fail' || result.Action === 'pass') {
        item.tests[foundIndex].pass = result.Action === 'pass'
        item.tests[foundIndex].completed = true
      }
    }

    // if package is a pass then we can remove it from the map since all tests passed
    if (
      (!result.Test &&
        (result.Action === 'pass' || result.Action === 'skip')) ||
      ('Output' in result && result.Output.includes('[no test files]'))
    ) {
      failureMap.delete(key)
    }

    // check for output with panics that have a test name and add it to a list of panic tests found
    if ('Output' in result && result.Output.includes('panic:')) {
      const pattern = /^panic:.* (Test[A-Z]\w*)/
      const match = result.Output.match(pattern)
      if (match) {
        const panicTestName = match[1]
        if (!item.panicTestsFound.includes(panicTestName)) {
          item.panicTestsFound.push(panicTestName)
        }
      }
    }
  }

  // now that we have all the packages and tests that failed we need to check for packages that failed
  // but no tests failed. This can happen when things like logging happen after a test failure
  for (const value of failureMap.values()) {
    let hasFailure = false
    for (const test of value.tests) {
      if (!test.pass) {
        hasFailure = true
      }
    }
    if (!hasFailure) {
      value.packageFailButNoTestFail = true
    }
  }

  return failureMap
}

/**
 * Read through the results file and write out the wanted failed test data
 * @param failureMap The map of test and package failures to write out
 * @param resultsFile The file to read through
 * @param outputToConsole Should we output to the console
 * @param outputFile A file we can optionally write out to the results to.
 * @returns A boolean for whether all tests and packages passed or not
 */
export async function processTestResultsFile(
  failureMap: Map<string, PassFailMapItem>,
  resultsFile: string,
  outputToConsole: boolean,
  outputFile: string
): Promise<boolean> {
  let outputStreamString = ''
  const stream = fs.createReadStream(resultsFile)
  const reader = readline.createInterface({input: stream})
  const packageWithoutTestFailuresHasBeenSeenBefore: string[] = []
  for await (const line of reader) {
    let result: TestResult
    try {
      const parsed = JSON.parse(line)
      result = TestResultSchema.parse(parsed)
    } catch (error) {
      // we ran into errors parsing the json, print it out because it may be useful for triage
      outputStreamString = streamOutput(
        outputStreamString,
        `${line}\n`,
        outputToConsole,
        outputFile
      )
      continue
    }

    // if we have package information then parse through it to set keys and such
    if (result.Package) {
      // if the failure map does not include this package then we can skip all logging for it
      const item = failureMap.get(result.Package)
      if (!item) {
        continue
      }

      // if we don't have output in this result then we can skip it.
      if (!('Output' in result)) {
        continue
      }

      // if the failure map has this package but no tests then we need to log all the output, weird case
      if (item.tests.length === 0 || item.packageFailButNoTestFail) {
        if (
          !packageWithoutTestFailuresHasBeenSeenBefore.includes(result.Package)
        ) {
          packageWithoutTestFailuresHasBeenSeenBefore.push(result.Package)
          outputStreamString = streamOutput(
            outputStreamString,
            `${result.Package} has failure logging but no test failures, the output below may be useful for triage\n`,
            outputToConsole,
            outputFile
          )
        }

        // if we have a panic test then we only want to log output for that test and not the others
        if (item.panicTestsFound.length > 0 && result.Test) {
          for (const panicTestName of item.panicTestsFound) {
            if (result.Test === panicTestName) {
              // Using a type assertion to reassure TypeScript about the type of `result`
              const output = (result as {Output: string}).Output
              outputStreamString = streamOutput(
                outputStreamString,
                output,
                outputToConsole,
                outputFile
              )
            }
          }
        } else {
          // Log out all other output and test logs
          outputStreamString = streamOutput(
            outputStreamString,
            result.Output,
            outputToConsole,
            outputFile
          )
        }
        continue
      }

      // if the failure map has this package and tests then we need to log the output for the tests that failed
      if (result.Test) {
        for (const test of item.tests) {
          // if this test failed then we log it
          if (test.test === result.Test && !test.pass) {
            // Using a type assertion to reassure TypeScript about the type of `result`
            // it loses track that we already checked for the existence of `Output` above
            // causing the static analysis to fail unless we have this line :shrug:
            const output = (result as {Output: string}).Output
            outputStreamString = streamOutput(
              outputStreamString,
              output,
              outputToConsole,
              outputFile
            )
          }
        }
      } else {
        // the package can have other fail data in it outside of a test such as a test logging after a failure
        // if the package data just says PASS then we can skip it since it can cause confusion
        if (result.Output === 'PASS\n') {
          continue
        }

        // add to output stream
        outputStreamString = streamOutput(
          outputStreamString,
          result.Output,
          outputToConsole,
          outputFile
        )
      }
    }
  }

  // write out the rest of the buffer before we exit
  flushStreamOutput(outputStreamString, outputToConsole, outputFile)
  return failureMap.size === 0
}

export function streamOutput(
  existingOutput: string,
  newOutput: string,
  shouldLog: boolean,
  outputFile: string
): string {
  existingOutput += newOutput
  // once the string gets to be pretty big lets flush it so we don't have to worry about memory issues
  if (existingOutput.length > 64000) {
    flushStreamOutput(existingOutput, shouldLog, outputFile)
    // reset existing output to empty since everything is now written out
    existingOutput = ''
  }
  return existingOutput
}
export function flushStreamOutput(
  existingOutput: string,
  shouldLog: boolean,
  outputFile: string
): void {
  // if nothing to flush then do nothing
  if (existingOutput.length === 0) {
    return
  }

  // only log if we want to
  if (shouldLog) {
    core.info(existingOutput)
  }

  // only output to file if we want to, heavily used in testing
  if (outputFile) {
    fs.appendFileSync(outputFile, existingOutput)
  }
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
