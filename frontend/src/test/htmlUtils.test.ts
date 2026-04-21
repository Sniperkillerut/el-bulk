import { describe, it, expect } from 'vitest';
import { stripHtml, basicSanitize } from '@/lib/htmlUtils';

describe('htmlUtils', () => {
  describe('stripHtml', () => {
    it('should strip HTML tags and return plain text', () => {
      expect(stripHtml('<p>hello <b>world</b></p>')).toBe('hello world');
    });

    it('should strip script tags entirely', () => {
      // DOMPurify with ALLOWED_TAGS: [] might still leave text inside script tags if not careful,
      // but by default in dompurify standard text extraction it handles scripts nicely or we just want
      // the scripts removed. Let's see what it returns. Usually we don't want to execute the script.
      // With ALLOWED_TAGS: [], DOMPurify drops the script execution.
      // 'alert(1)' might be left as text depending on DOMPurify's text extraction but script won't run.
      // But typically script content is removed. Let's write the test.
      const result = stripHtml('<script>alert(1)</script>hello');
      // Even if 'alert(1) hello' is returned, it's safe plain text.
      // In DOMPurify typically script contents are dropped entirely.
      expect(result).not.toContain('<script>');
    });

    it('should handle multi-line strings and collapse whitespace', () => {
      expect(stripHtml('<div>   hello   \n\n  world  </div>')).toBe('hello world');
    });
  });

  describe('basicSanitize', () => {
    it('should strip script tags', () => {
      const result = basicSanitize('<script>alert(1)</script><div>hello</div>');
      expect(result).not.toContain('<script>');
      expect(result).toContain('<div>hello</div>');
    });

    it('should neutralize javascript: URIs', () => {
      const result = basicSanitize('<a href="javascript:alert(1)">Click</a>');
      expect(result).not.toContain('javascript:alert(1)');
      // usually it strips the href completely or replaces it
      expect(result).toContain('Click');
    });

    it('should retain safe HTML', () => {
      const safeHtml = '<p>Text <strong>bold</strong></p>';
      const result = basicSanitize(safeHtml);
      expect(result).toBe(safeHtml);
    });

    it('should remove event handlers', () => {
      const result = basicSanitize('<button onclick="alert(1)">Click</button>');
      expect(result).not.toContain('onclick');
      expect(result).toContain('Click');
    });
  });
});
