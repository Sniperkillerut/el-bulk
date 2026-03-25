'use client';

import { useEffect } from 'react';
import { remoteLogger } from '@/lib/remoteLogger';

export default function RemoteLogManager() {
  useEffect(() => {
    const handleError = (event: ErrorEvent) => {
      // Prevent duplicate logging for errors already captured and tagged
      if (event.error?._remoteLogged) return;

      remoteLogger.error(`Unhandled Error: ${event.message}`, {
        filename: event.filename,
        lineno: event.lineno,
        colno: event.colno,
        error: event.error?.stack || event.error?.message,
      });
    };

    const handleRejection = (event: PromiseRejectionEvent) => {
      // Prevent duplicate logging for rejections already captured and tagged
      if (event.reason?._remoteLogged) return;

      remoteLogger.error(`Unhandled Promise Rejection: ${event.reason?.message || event.reason}`, {
        reason: event.reason?.stack || event.reason,
      });
    };

    window.addEventListener('error', handleError);
    window.addEventListener('unhandledrejection', handleRejection);

    // Optional: Log when the app session starts
    remoteLogger.debug('Frontend RemoteLogManager initialized');

    return () => {
      window.removeEventListener('error', handleError);
      window.removeEventListener('unhandledrejection', handleRejection);
    };
  }, []);

  return null;
}
