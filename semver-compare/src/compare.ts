import semver from 'semver'

export type Operator = 'gt' | 'lt' | 'eq'

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
