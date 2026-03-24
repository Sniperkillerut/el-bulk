'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';

export default function OrderConfirmation() {
  const params = useParams();
  const orderNumber = params.order_number as string;

  return (
    <div className="centered-container px-4 py-16 text-center">
      <div className="card no-tilt max-w-lg mx-auto p-8">
        <div className="text-5xl mb-4">✅</div>
        <h1 className="font-display text-4xl mb-2">¡ORDEN RECIBIDA!</h1>
        <div className="divider" />

        <div className="cardbox p-4 mb-6" style={{ background: 'var(--kraft-light)' }}>
          <p className="text-xs font-mono-stack mb-1" style={{ color: 'var(--text-muted)' }}>NÚMERO DE ORDEN</p>
          <p className="font-display text-3xl" style={{ color: 'var(--gold-dark)' }}>{orderNumber}</p>
        </div>

        <p className="text-sm mb-6" style={{ color: 'var(--text-secondary)' }}>
          Tu pedido ha sido registrado exitosamente. Un asesor se pondrá en contacto contigo para coordinar el pago y la entrega.
        </p>

        <div className="flex flex-col sm:flex-row gap-3 justify-center">
          <Link href="/" className="btn-primary">← VOLVER A LA TIENDA</Link>
          <Link href="/contact" className="btn-secondary">CONTACTAR</Link>
        </div>
      </div>
    </div>
  );
}
