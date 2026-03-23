-- Migration: Add is_active column to custom_categories

ALTER TABLE custom_categories 
ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;
