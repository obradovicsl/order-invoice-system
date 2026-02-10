import { describe, it, expect } from 'vitest';
import { cn } from './utils';

describe('cn utility function', () => {
  it('should combine class names correctly', () => {
    const result = cn('px-2', 'py-1', 'bg-red-500');
    expect(result).toContain('px-2');
    expect(result).toContain('py-1');
    expect(result).toContain('bg-red-500');
  });

  it('should handle tailwind class merging and override', () => {
    const result = cn('px-2 py-1', 'px-4');
    // px-4 should override px-2
    expect(result).toContain('px-4');
    expect(result).not.toMatch(/px-2(?!\d)/);
  });

  it('should handle conditional classes with objects', () => {
    const isActive = true;
    const isDisabled = false;
    const result = cn({
      'bg-blue-500': isActive,
      'bg-gray-300': isDisabled,
    });
    expect(result).toContain('bg-blue-500');
    expect(result).not.toContain('bg-gray-300');
  });

  it('should handle empty input', () => {
    const result = cn('');
    expect(result).toBeDefined();
    expect(typeof result).toBe('string');
  });

  it('should handle multiple arguments with mixed types', () => {
    const result = cn('px-2', 'py-1', undefined, null, 'bg-white');
    expect(result).toContain('px-2');
    expect(result).toContain('py-1');
    expect(result).toContain('bg-white');
  });
});
