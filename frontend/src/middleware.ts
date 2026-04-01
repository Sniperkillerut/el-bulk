import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  const adminToken = request.cookies.get('admin_token')?.value;
  const { pathname } = request.nextUrl;

  // Protect /admin routes (except login)
  if (pathname.startsWith('/admin') && pathname !== '/admin/login') {
    if (!adminToken) {
      const loginUrl = new URL('/admin/login', request.url);
      // Optional: Add redirect parameter to return here after login
      // loginUrl.searchParams.set('callbackUrl', pathname);
      return NextResponse.redirect(loginUrl);
    }
  }

  return NextResponse.next();
}

// Ensure middleware only runs on relevant routes to save performance
export const config = {
  matcher: ['/admin/:path*'],
};
