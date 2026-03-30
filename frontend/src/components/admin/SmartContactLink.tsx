'use client';

/**
 * SmartContactLink detects if a contact string is an email or a phone number
 * and returns the appropriate link (mailto: or https://wa.me/).
 */
export default function SmartContactLink({ contact, className }: { contact: string; className?: string }) {
  if (!contact) return <span className={className}>—</span>;
  
  // Basic detection: if it contains '@', it's an email.
  const isEmail = contact.includes('@');
  
  if (isEmail) {
    return (
      <a 
        href={`mailto:${contact}`} 
        className={className}
        target="_blank" 
        rel="noopener noreferrer"
      >
        {contact}
      </a>
    );
  }
  
  // Otherwise, assume it's a phone number and link to WhatsApp.
  const cleanPhone = contact.replace(/\D/g, '');
  
  // If the phone number is too short or invalid after cleaning, just show text.
  if (cleanPhone.length < 5) {
     return <span className={className}>{contact}</span>;
  }

  return (
    <a 
      href={`https://wa.me/${cleanPhone}`} 
      className={className}
      target="_blank" 
      rel="noopener noreferrer"
    >
      {contact}
    </a>
  );
}
