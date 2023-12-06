import {z} from 'zod'

/**
 * The subschema for the test results that we want to handle
 * when generating the test result summary
 */
const handledTestStatuses = z.enum(['pass', 'fail'])
export type HandledTestResults = z.infer<typeof handledTestResultsSchema>
export const handledTestResultsSchema = z.object({
  Package: z.string().min(1).optional(),
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
export const TestResultSchema = z.discriminatedUnion('Action', [
  handledTestResultsSchema,
  z.object({
    Test: z.string().min(1).optional(),
    Action: z.literal('skip'),
    Elapsed: z.number().nonnegative(),
    Package: z.string().min(1).optional()
  }),
  z.object({
    Test: z.string().min(1),
    Action: z.literal('run'),
    Package: z.string().min(1).optional()
  }),
  z.object({
    Output: z.string().min(1),
    Action: z.literal('output'),
    Test: z.string().min(1).optional(),
    Package: z.string().min(1).optional()
  }),
  z.object({
    Action: z.literal('start'),
    Package: z.string().min(1).optional(),
    Test: z.string().min(1).optional()
  }),
  z.object({
    Action: z.literal('pause'),
    Package: z.string().min(1).optional(),
    Test: z.string().min(1).optional()
  }),
  z.object({
    Action: z.literal('cont'),
    Package: z.string().min(1).optional(),
    Test: z.string().min(1).optional()
  })
])
/**
 * A representation of a file containing all test logs of
 * the JSONL format outputted by https://pkg.go.dev/cmd/test2json
 */
export const TestResultsSchema = z.array(TestResultSchema)

const outputMode = z.enum(['standard', 'json'])
export type OutputMode = z.infer<typeof outputMode>

export type PassFailTest = z.infer<typeof passFailTestSchema>
export const passFailTestSchema = z.object({
  test: z.string().min(1),
  pass: z.boolean(),
  completed: z.boolean()
})

export type PassFailMapItem = z.infer<typeof passFailMapSchema>
export const passFailMapSchema = z.object({
  tests: z.array(passFailTestSchema),
  packageFailButNoTestFail: z.boolean(),
  panicTestsFound: z.array(z.string().min(1))
})
