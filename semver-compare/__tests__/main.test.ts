import {versionCompare} from '../src/compare'
import semver from 'semver'

describe('versionCompare', () => {
  test('should throw an error when no versions or operator specified', () => {
    expect(() => versionCompare('', 'gt', '')).toThrow(
      'Required inputs not specified.'
    )
    expect(() => versionCompare('1.0.0', 'gt', '')).toThrow(
      'Required inputs not specified.'
    )
    expect(() => versionCompare('', 'gt', '1.0.0')).toThrow(
      'Required inputs not specified.'
    )
  })

  test('should throw an error when versions are invalid', () => {
    expect(() => versionCompare('abc', 'gt', '1.0.0')).toThrow(
      'Invalid version(s).'
    )
    expect(() => versionCompare('1.0.0', 'gt', 'abc')).toThrow(
      'Invalid version(s).'
    )
  })

  test('should throw an error when operator is invalid', () => {
    expect(() => versionCompare('1.0.0', 'abc' as any, '1.0.0')).toThrow(
      'Invalid operator.'
    )
  })

  test('should return the correct result', () => {
    expect(versionCompare('1.0.0', 'gt', '0.1.0')).toBe(true)
    expect(versionCompare('1.0.0', 'lt', '2.0.0')).toBe(true)
    expect(versionCompare('1.0.0', 'eq', '1.0.0')).toBe(true)
  })
})
