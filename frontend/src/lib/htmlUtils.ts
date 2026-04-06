/**
 * Utility for safe HTML processing using browser-native DOMParser.
 *
 * NOTE: While DOMPurify is the preferred industry standard for HTML sanitization,
 * it is currently unavailable due to environment constraints. These utilities
 * provide a robust alternative using browser-native APIs and careful fallbacks.
 */

/**
 * Strips all HTML tags and returns plain text content.
 * Safe for use when you need to render a snippet or preview as plain text.
 */
export function stripHtml(html: string): string {
  if (typeof window === 'undefined') {
    // Server-side fallback: strip tags with regex. This is safe as long as the
    // result is rendered as plain text in React, which escapes it automatically.
    return html.replace(/<[^>]*>?/gm, ' ').replace(/\s+/g, ' ').trim();
  }

  try {
    const parser = new DOMParser();
    const doc = parser.parseFromString(html, 'text/html');
    return (doc.body.textContent || doc.body.innerText || '').replace(/\s+/g, ' ').trim();
  } catch (err) {
    // Fallback if parsing fails
    return html.replace(/<[^>]*>?/gm, ' ').replace(/\s+/g, ' ').trim();
  }
}

/**
 * Performs basic sanitization on HTML content by removing known dangerous tags
 * and event handlers. This uses DOMParser on the client and a multi-pass regex
 * approach on the server.
 */
export function basicSanitize(html: string): string {
  if (typeof window === 'undefined') {
    return basicSanitizeSSRFallback(html);
  }

  try {
    const parser = new DOMParser();
    const doc = parser.parseFromString(html, 'text/html');

    // 1. Remove dangerous tags
    const dangerousTags = doc.querySelectorAll('script, object, embed, iframe, form, button, meta, link, style');
    dangerousTags.forEach(t => t.remove());

    // 2. Remove on* event handler attributes and javascript: URIs from all elements
    const allElements = doc.querySelectorAll('*');
    allElements.forEach(el => {
      const attributes = Array.from(el.attributes);
      attributes.forEach(attr => {
        const name = attr.name.toLowerCase();
        const value = attr.value.toLowerCase().trim();

        if (name.startsWith('on')) {
          el.removeAttribute(attr.name);
        } else if ((name === 'href' || name === 'src' || name === 'action' || name === 'formaction') &&
                   value.startsWith('javascript:')) {
          el.removeAttribute(attr.name);
          // Set safe fallback for links
          if (name === 'href') el.setAttribute('href', '#');
        }
      });
    });

    return doc.body.innerHTML;
  } catch (err) {
    // If browser parsing fails, fallback to the same regex logic used for SSR
    return basicSanitizeSSRFallback(html);
  }
}

/**
 * Shared regex-based sanitization for server-side and browser fallback.
 * Separated for clarity and testing.
 */
function basicSanitizeSSRFallback(html: string): string {
  let sanitized = html;

  // 1. Strip <script>...</script>
  sanitized = sanitized.replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '');

  // 2. Strip dangerous tags
  sanitized = sanitized.replace(/<(object|embed|iframe|form|button|meta|link|style)\b[^>]*>([\s\S]*?<\/\1>)?/gi, '');

  // 3. Strip all 'on*' attributes (event handlers)
  sanitized = sanitized.replace(/\son[a-z]+\s*=\s*("[^"]*"|'[^']*'|[^\s>]*)/gi, '');

  // 4. Strip 'javascript:' URI protocol in common attributes
  sanitized = sanitized.replace(/\s(href|src|action|formaction)\s*=\s*("|')\s*javascript:[^"']*\2/gi, ' $1="#"');

  return sanitized;
}
