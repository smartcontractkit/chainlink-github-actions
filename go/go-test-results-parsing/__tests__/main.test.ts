import {
  deserializeTestResultsFile,
  jsonTestOutput,
  getTestFailures,
  standardTestOutput
} from '../src/main'
import {TestResult} from '../src/types'
const file = './__tests__/fixtures/go_test_results_input.json'
describe('output testing', () => {
  test('can deserialize file', () => {
    const testResults: TestResult[] = deserializeTestResultsFile(file)
    expect(testResults).toMatchSnapshot()
    expect(testResults.length).toBe(11)
  })
  test('failed test list', () => {
    const testResults: TestResult[] = deserializeTestResultsFile(file)
    const failedTests: string[] = getTestFailures(testResults)
    expect(failedTests).toMatchSnapshot()
    expect(failedTests.length).toBe(1)
  })
  test('standard test output', () => {
    const testResults: TestResult[] = deserializeTestResultsFile(file)
    const failedTests: string[] = getTestFailures(testResults)
    const filteredResults: TestResult[] = standardTestOutput(
      failedTests,
      testResults,
      true
    )
    expect(filteredResults).toMatchSnapshot()
    expect(filteredResults.length).toBe(1)
  })
  test('json test output', () => {
    const testResults: TestResult[] = deserializeTestResultsFile(file)
    const failedTests: string[] = getTestFailures(testResults)
    const filteredResults: TestResult[] = jsonTestOutput(
      failedTests,
      testResults,
      true
    )
    expect(filteredResults).toMatchSnapshot()
    expect(filteredResults.length).toBe(5)
  })
})
