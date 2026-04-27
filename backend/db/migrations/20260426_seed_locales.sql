-- Seed file for locale metadata translations
INSERT INTO translation (key, locale, value) VALUES
-- English labels
('system.locale.en.label', 'en', 'English'),
('system.locale.en.icon', 'en', '🇺🇸'),
('system.locale.es.label', 'en', 'Spanish'),
('system.locale.es.icon', 'en', '🇪🇸'),
('system.locale.fr.label', 'en', 'French'),
('system.locale.fr.icon', 'en', '🇫🇷'),
('system.locale.pt.label', 'en', 'Portuguese'),
('system.locale.pt.icon', 'en', '🇵🇹'),
('system.locale.de.label', 'en', 'German'),
('system.locale.de.icon', 'en', '🇩🇪'),
('system.locale.it.label', 'en', 'Italian'),
('system.locale.it.icon', 'en', '🇮🇹'),

-- Spanish labels
('system.locale.en.label', 'es', 'Inglés'),
('system.locale.en.icon', 'es', '🇺🇸'),
('system.locale.es.label', 'es', 'Español'),
('system.locale.es.icon', 'es', '🇪🇸'),
('system.locale.fr.label', 'es', 'Francés'),
('system.locale.fr.icon', 'es', '🇫🇷'),
('system.locale.pt.label', 'es', 'Portugués'),
('system.locale.pt.icon', 'es', '🇵🇹'),
('system.locale.de.label', 'es', 'Alemán'),
('system.locale.de.icon', 'es', '🇩🇪'),
('system.locale.it.label', 'es', 'Italiano'),
('system.locale.it.icon', 'es', '🇮🇹'),

-- Native labels (Self-description)
('system.locale.fr.label', 'fr', 'Français'),
('system.locale.fr.icon', 'fr', '🇫🇷'),
('system.locale.pt.label', 'pt', 'Português'),
('system.locale.pt.icon', 'pt', '🇵🇹'),
('system.locale.de.label', 'de', 'Deutsch'),
('system.locale.de.icon', 'de', '🇩🇪'),
('system.locale.it.label', 'it', 'Italiano'),
('system.locale.it.icon', 'it', '🇮🇹')
ON CONFLICT (key, locale) DO UPDATE SET value = EXCLUDED.value;
