import * as fs from 'fs'
import * as path from 'path'
import {preprocessTestResultsFile, processTestResultsFile} from '../src/main'
import {PassFailMapItem} from '../src/types'

const fileBase = './__tests__/fixtures/'
const tmpFile = `${fileBase}tmp_test_`

function cleanUpTestFiles(): void {
  const files = fs.readdirSync(fileBase)
  files.forEach(file => {
    if (file.startsWith('tmp_test_')) {
      const filePath = path.join(fileBase, file)

      try {
        fs.unlinkSync(filePath)
      } catch (unlinkErr) {
        console.error('Error deleting file:', unlinkErr)
      }
    }
  })
}

describe('output testing', () => {
  // cleanup test files before and after test
  beforeAll(() => {
    cleanUpTestFiles()
  })
  afterAll(() => {
    cleanUpTestFiles()
  })

  it('can read file with all passes', async () => {
    const outputFile = `${tmpFile}all_passing.txt`
    const inputFile = `${fileBase}all_passing.json`
    const failureMap: Map<string, PassFailMapItem> =
      await preprocessTestResultsFile(inputFile)
    const pass: boolean = await processTestResultsFile(
      failureMap,
      inputFile,
      false,
      outputFile
    )
    expect(pass).toBe(true)
    expect(fs.existsSync(outputFile)).toBe(false)
  })
  it('can read file with all failures', async () => {
    const outputFile = `${tmpFile}all_failing.txt`
    const inputFile = `${fileBase}all_failing.json`
    const failureMap: Map<string, PassFailMapItem> =
      await preprocessTestResultsFile(inputFile)
    const pass: boolean = await processTestResultsFile(
      failureMap,
      inputFile,
      false,
      outputFile
    )
    expect(pass).toBe(false)
    const fileContent = fs.readFileSync(outputFile, 'utf-8')
    expect(fileContent).toMatchSnapshot()
  })
  it('can read file with mix of pass and failures', async () => {
    const outputFile = `${tmpFile}pass_fail_mix.txt`
    const inputFile = `${fileBase}pass_fail_mix.json`
    const failureMap: Map<string, PassFailMapItem> =
      await preprocessTestResultsFile(inputFile)
    const pass: boolean = await processTestResultsFile(
      failureMap,
      inputFile,
      false,
      outputFile
    )
    expect(pass).toBe(false)
    const fileContent = fs.readFileSync(outputFile, 'utf-8')
    expect(fileContent).toMatchSnapshot()
  })
  it('can read file with mix of pass and failures failures and non json txt injected from other potential errors in the go runner', async () => {
    const outputFile = `${tmpFile}pass_fail_mix_with_non_json.txt`
    const inputFile = `${fileBase}pass_fail_mix_with_non_json.json`
    const failureMap: Map<string, PassFailMapItem> =
      await preprocessTestResultsFile(inputFile)
    const pass: boolean = await processTestResultsFile(
      failureMap,
      inputFile,
      false,
      outputFile
    )
    expect(pass).toBe(false)
    const fileContent = fs.readFileSync(outputFile, 'utf-8')
    expect(fileContent).toMatchSnapshot()
  })

  it('can parse a panic for a test name and only output logs for that panic from that package', async () => {
    const outputFile = `${tmpFile}package_panic_with_test_name.txt`
    const inputFile = `${fileBase}package_panic_with_test_name.json`
    const failureMap: Map<string, PassFailMapItem> =
      await preprocessTestResultsFile(inputFile)
    const pass: boolean = await processTestResultsFile(
      failureMap,
      inputFile,
      false,
      outputFile
    )
    expect(pass).toBe(false)
    const fileContent = fs.readFileSync(outputFile, 'utf-8')
    expect(fileContent).toMatchSnapshot()
  })
  it('can read file with package failure but no test failure', async () => {
    const outputFile = `${tmpFile}failing_package_without_test_fail.txt`
    const inputFile = `${fileBase}failing_package_without_test_fail.json`
    const failureMap: Map<string, PassFailMapItem> =
      await preprocessTestResultsFile(inputFile)
    const pass: boolean = await processTestResultsFile(
      failureMap,
      inputFile,
      false,
      outputFile
    )
    expect(pass).toBe(false)
    const fileContent = fs.readFileSync(outputFile, 'utf-8')
    expect(fileContent).toMatchSnapshot()
  })
})
