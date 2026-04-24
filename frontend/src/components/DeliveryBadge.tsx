'use client';

import React, { useEffect, useState } from 'react';
import { useLanguage } from '@/context/LanguageContext';
import { fetchPublicSettings } from '@/lib/api';
import { PublicSettings } from '@/lib/types';

const DeliveryBadge: React.FC = () => {
  const { t } = useLanguage();
  const [settings, setSettings] = useState<PublicSettings | null>(null);
  const [isOpen, setIsOpen] = useState(false);

  const checkIfOpen = (hoursStr: string): boolean => {
    try {
      // Basic parser for "Mon - Sat: 11:00 AM - 7:00 PM"
      const now = new Date();
      const bogotaTime = new Date(now.toLocaleString("en-US", {timeZone: "America/Bogota"}));
      const day = bogotaTime.getDay(); // 0=Sun, 1=Mon, ..., 6=Sat
      
      // Check if it's Sunday (Store usually closed on Sun based on "Mon - Sat")
      if (day === 0 && !hoursStr.toLowerCase().includes('sun')) return false;

      // Extract time part: "11:00 AM - 7:00 PM"
      let timePart = hoursStr;
      if (hoursStr.includes(':')) {
         const parts = hoursStr.split(':');
         // Join parts after the first colon in case there are multiple
         timePart = parts.slice(1).join(':').trim();
      }
      
      const [startStr, endStr] = timePart.split('-').map(s => s.trim());
      
      const parseTime = (str: string) => {
        const [time, modifier] = str.split(' ');
        const [hours, minutesStr] = time.split(':');
        let hoursNum = Number(hours);
        const minutes = Number(minutesStr) || 0;
        
        if (modifier === 'PM' && hoursNum < 12) hoursNum += 12;
        if (modifier === 'AM' && hoursNum === 12) hoursNum = 0;
        return hoursNum * 60 + minutes;
      };

      const currentMinutes = bogotaTime.getHours() * 60 + bogotaTime.getMinutes();
      const startMinutes = parseTime(startStr);
      const endMinutes = parseTime(endStr);

      return currentMinutes >= startMinutes && currentMinutes <= endMinutes;
    } catch (e) {
      console.warn('Failed to parse delivery hours', e);
      return true; // Fallback to true if manual override is enabled but hours are weird
    }
  };

  useEffect(() => {
    const loadSettings = async () => {
      try {
        const data = await fetchPublicSettings({ forceRefresh: true });
        setSettings(data);
        
        if (data.delivery_priority_enabled && data.contact_hours) {
          setIsOpen(checkIfOpen(data.contact_hours));
        } else {
          setIsOpen(false);
        }
      } catch (err) {
        console.error('Failed to load settings for DeliveryBadge', err);
      }
    };

    loadSettings();
    // Refresh every 2 minutes
    const interval = setInterval(loadSettings, 2 * 60 * 1000);
    return () => clearInterval(interval);
  }, []);

  const isActive = settings?.delivery_priority_enabled && isOpen;

  return (
    <div className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-sm border text-[10px] font-mono-stack font-bold tracking-wider transition-all duration-500 ${
      isActive 
        ? 'bg-green-500/10 border-green-500/50 text-green-400 shadow-[0_0_15px_rgba(34,197,94,0.1)]' 
        : 'bg-white/5 border-white/10 text-white/40'
    }`}>
      <span className={`relative flex h-2 w-2`}>
        {isActive && (
          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
        )}
        <span className={`relative inline-flex rounded-full h-2 w-2 ${isActive ? 'bg-green-500' : 'bg-white/20'}`}></span>
      </span>
      <span>
        {isActive 
          ? t('components.delivery.available', 'BOGOTÁ EXPRESS: AVAILABLE NOW ⚡')
          : t('components.delivery.offline', 'BOGOTÁ EXPRESS: CURRENTLY OFFLINE 📦')
        }
      </span>
    </div>
  );
};

export default DeliveryBadge;
