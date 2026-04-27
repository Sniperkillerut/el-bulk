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
      if (!hoursStr || !hoursStr.startsWith('{')) return false;

      const now = new Date();
      const bogotaTime = new Date(now.toLocaleString("en-US", { timeZone: "America/Bogota" }));
      const day = bogotaTime.getDay(); // 0=Sun, 1=Mon, ..., 6=Sat
      const days = ['sun', 'mon', 'tue', 'wed', 'thu', 'fri', 'sat'];
      const today = days[day];
      
      const currentMinutes = bogotaTime.getHours() * 60 + bogotaTime.getMinutes();

      const config = JSON.parse(hoursStr);
      const dayConfig = config[today];
      
      if (!dayConfig || !dayConfig.open) return false;
      
      const parseHHmm = (str: string) => {
        const [h, m] = str.split(':').map(Number);
        return h * 60 + m;
      };
      
      const startMin = parseHHmm(dayConfig.start);
      const endMin = parseHHmm(dayConfig.end);
      
      if (startMin > endMin) { // Overnight shift support
        return currentMinutes >= startMin || currentMinutes <= endMin;
      }
      return currentMinutes >= startMin && currentMinutes <= endMin;
    } catch (e) {
      console.warn('Failed to parse delivery hours JSON', e);
      return false;
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
    // Refresh every 30 seconds for better responsiveness during testing
    const interval = setInterval(loadSettings, 30 * 1000);
    return () => clearInterval(interval);
  }, []);

  const isActive = settings?.delivery_priority_enabled && isOpen;

  return (
    <div className={`inline-flex items-center gap-2 px-3 py-1 rounded-none border font-mono-stack font-bold tracking-widest transition-all duration-300 select-none ${
      isActive 
        ? 'bg-[#DCFCE7] border-[#86EFAC] text-[#166534] shadow-sm' 
        : 'bg-[#FDFBF7]/80 border-[#E6DAC3] text-[#8B795C] backdrop-blur-sm'
    }`}>
      <div className="relative flex h-2 w-2">
        {isActive && (
          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-[#166534] opacity-20"></span>
        )}
        <div className={`relative inline-flex rounded-none h-2 w-2 border ${isActive ? 'bg-[#166534] border-[#166534]' : 'bg-[#8B795C]/20 border-[#8B795C]/30'}`}></div>
      </div>
      <span className="text-[10px] uppercase whitespace-nowrap">
        {isActive 
          ? t('components.delivery.available', 'BOGOTÁ EXPRESS: ACTIVE ⚡')
          : t('components.delivery.offline', 'BOGOTÁ EXPRESS: INACTIVE 📦')
        }
      </span>
    </div>
  );
};

export default DeliveryBadge;
