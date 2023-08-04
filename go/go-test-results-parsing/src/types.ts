import {z} from 'zod'

/**
 * The subschema for the test results that we want to handle
 * when generating the test result summary
 */
const handledTestStatuses = z.enum(['pass', 'fail'])
export type HandledTestResults = z.infer<typeof handledTestResultsSchema>
export const handledTestResultsSchema = z.object({
  Test: z.string().min(1).optional(),
  Action: handledTestStatuses,
  Elapsed: z.number().nonnegative()
})

/**
 * Possible test result actions formats
 *
 * Ideally, we would use `transform` to map the parsed test result to the
 * fields we want to use in the test result summary, but that breaks
 * discriminated unions.
 *
 * @see https://github.com/colinhacks/zod/issues/2315
 */
export type TestResult = z.infer<typeof TestResultSchema>
const TestResultSchema = z.discriminatedUnion('Action', [
  handledTestResultsSchema,
  z.object({
    Test: z.string().min(1).optional(),
    Action: z.literal('skip'),
    Elapsed: z.number().nonnegative()
  }),
  z.object({
    Test: z.string().min(1),
    Action: z.literal('run')
  }),
  z.object({
    Output: z.string().min(1),
    Action: z.literal('output'),
    Test: z.string().min(1).optional()
  }),
  z.object({
    Action: z.literal('start')
  }),
  z.object({
    Action: z.literal('pause')
  }),
  z.object({
    Action: z.literal('cont')
  })
])
/**
 * A representation of a file containing all test logs of
 * the JSONL format outputted by https://pkg.go.dev/cmd/test2json
 */
export const TestResultsSchema = z.array(TestResultSchema)

const outputMode = z.enum(['standard', 'json'])
export type OutputMode = z.infer<typeof outputMode>

const testRunFailures = z.object({
  TestsFailed: z.array(z.string()),
  PackageFailure: z.boolean()
})
export type TestRunFailures = z.infer<typeof testRunFailures>
