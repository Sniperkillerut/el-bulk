import DOMPurify from 'isomorphic-dompurify';

/**
 * Utility for safe HTML processing using isomorphic-dompurify.
 */

/**
 * Strips all HTML tags and returns plain text content.
 * Safe for use when you need to render a snippet or preview as plain text.
 */
export function stripHtml(html: string): string {
  // Use DOMPurify to strip all tags.
  // Note: DOMPurify's ALLOWED_TAGS: [] keeps the text content.
  // We still do a replace to collapse multiple spaces/newlines into a single space,
  // to maintain parity with the previous implementation's return value.
  const purified = DOMPurify.sanitize(html, { ALLOWED_TAGS: [] });
  return purified.replace(/\s+/g, ' ').trim();
}

/**
 * Performs basic sanitization on HTML content by removing known dangerous tags
 * and event handlers.
 */
export function basicSanitize(html: string): string {
  // Use DOMPurify with the default html profile which allows standard safe HTML
  // but strips scripts, objects, and event handlers.
  return DOMPurify.sanitize(html, { USE_PROFILES: { html: true } });
}
