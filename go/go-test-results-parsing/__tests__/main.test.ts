import {
  deserializeTestResultsFile,
  jsonTestOutput,
  getTestFailures,
  standardTestOutput,
  logAllOutput
} from '../src/main'
import {TestResult, TestRunFailures} from '../src/types'
const file = './__tests__/fixtures/go_test_results_input.json'
describe('output testing', () => {
  test('can deserialize file', () => {
    const testResults: TestResult[] = deserializeTestResultsFile(file)
    expect(testResults).toMatchSnapshot()
    expect(testResults.length).toBe(11)
  })
  test('failed test list', () => {
    const testResults: TestResult[] = deserializeTestResultsFile(file)
    const failedTests: TestRunFailures = getTestFailures(testResults)
    expect(failedTests.TestsFailed).toMatchSnapshot()
    expect(failedTests.TestsFailed.length).toBe(1)
    expect(failedTests.PackageFailure).toBe(true)
  })
  test('standard test output', () => {
    const testResults: TestResult[] = deserializeTestResultsFile(file)
    const failedTests: TestRunFailures = getTestFailures(testResults)
    const filteredResults: TestResult[] = standardTestOutput(
      failedTests.TestsFailed,
      testResults,
      true
    )
    expect(filteredResults).toMatchSnapshot()
    expect(filteredResults.length).toBe(1)
    expect(failedTests.PackageFailure).toBe(true)
  })
  test('json test output', () => {
    const testResults: TestResult[] = deserializeTestResultsFile(file)
    const failedTests: TestRunFailures = getTestFailures(testResults)
    const filteredResults: TestResult[] = jsonTestOutput(
      failedTests.TestsFailed,
      testResults,
      true
    )
    expect(filteredResults).toMatchSnapshot()
    expect(filteredResults.length).toBe(5)
    expect(failedTests.PackageFailure).toBe(true)
  })
  test('failure without a test failure', () => {
    const file = './__tests__/fixtures/failure_output_without_test_failure.json'
    const testResults: TestResult[] = deserializeTestResultsFile(file)
    const failedTests: TestRunFailures = getTestFailures(testResults)
    const filteredResults: TestResult[] = standardTestOutput(
      failedTests.TestsFailed,
      testResults,
      true
    )
    expect(filteredResults).toMatchSnapshot()
    expect(filteredResults.length).toBe(0)
    expect(failedTests.PackageFailure).toBe(true)
    logAllOutput(testResults)
  })
})
