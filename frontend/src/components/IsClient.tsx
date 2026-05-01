'use client';

import { ReactNode, useEffect, useState } from 'react';

interface IsClientProps {
  children: ReactNode;
  fallback?: ReactNode;
}

/**
 * A utility component that only renders its children on the client-side.
 * This is useful for avoiding hydration mismatches when rendering content
 * that depends on browser-only globals (window, localStorage, etc.) 
 * or dynamic values that might differ from the server render.
 */
export default function IsClient({ children, fallback = null }: IsClientProps) {
  const [isClient, setIsClient] = useState(false);

  useEffect(() => {
    setIsClient(true);
  }, []);

  if (!isClient) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
}
