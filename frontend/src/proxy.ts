import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export default function proxy(request: NextRequest) {
  const adminToken = request.cookies.get('admin_token')?.value;
  const userToken = request.cookies.get('user_token')?.value;
  const { pathname } = request.nextUrl;

  // 1. Admin Protection
  if (pathname.startsWith('/admin')) {
    // Skip protection for the login page itself to avoid infinite loops
    if (pathname === '/admin/login') {
      // If already logged in, redirect to dashboard
      if (adminToken) {
        return NextResponse.redirect(new URL('/admin/dashboard', request.url));
      }
      return NextResponse.next();
    }

    // Protect all other admin routes
    if (!adminToken) {
      return NextResponse.redirect(new URL('/admin/login', request.url));
    }
  }

  // 2. User Protection
  const protectedUserRoutes = ['/checkout', '/orders'];
  const isProtectedRoute = protectedUserRoutes.some(route => pathname.startsWith(route));

  if (isProtectedRoute && !userToken) {
    const loginUrl = new URL('/login', request.url);
    loginUrl.searchParams.set('callbackUrl', pathname + request.nextUrl.search);
    return NextResponse.redirect(loginUrl);
  }

  // Allow access
  return NextResponse.next();
}

// Ensure middleware only runs on relevant routes to save performance
export const config = {
  matcher: ['/admin/:path*', '/checkout/:path*', '/orders/:path*'],
};
