'use client';

import { useState } from 'react';

interface ContactForm {
  customer_name: string;
  customer_contact: string;
  [key: string]: any;
}

export function useForm<T extends ContactForm>(initialState: T) {
  const [form, setForm] = useState<T>(initialState);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;
    setForm(prev => ({
      ...prev,
      [name]: type === 'number' ? parseInt(value) || 0 : value
    }));
  };

  const setFieldValue = (name: string, value: any) => {
    setForm(prev => ({ ...prev, [name]: value }));
  };

  const validate = () => {
    if (!form.customer_name || !form.customer_contact) {
      setError('Please provide your name and contact info.');
      return false;
    }
    return true;
  };

  const handleSubmit = async (onSubmit: (data: T) => Promise<void>) => {
    if (!validate()) return;

    setSubmitting(true);
    setError('');

    try {
      await onSubmit(form);
      setSuccess(true);
    } catch (err: unknown) {
      if (err instanceof Error && err.message) {
        setError(err.message);
      } else {
        setError('An error occurred. Please try again.');
      }
    } finally {
      setSubmitting(false);
    }
  };

  return {
    form,
    setForm,
    handleChange,
    setFieldValue,
    submitting,
    error,
    setError,
    success,
    setSuccess,
    handleSubmit
  };
}
