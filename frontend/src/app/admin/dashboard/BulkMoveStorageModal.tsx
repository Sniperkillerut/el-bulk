'use client';

import React, { useState, useEffect, useCallback } from 'react';
import Modal from '@/components/ui/Modal';
import Image from 'next/image';
import { useLanguage } from '@/context/LanguageContext';
import { Product, StorageLocation, StoredIn } from '@/lib/types';
import { adminFetchStorage, adminBulkMoveStorage } from '@/lib/api';
import { useToast } from '@/context/ToastContext';
import { resolveLabel, FOIL_LABELS, TREATMENT_LABELS } from '@/lib/types';

interface MoveItem {
  product: Product;
  sourceLocation: StorageLocation;
  currentQuantity: number;
}

interface BulkMoveStorageModalProps {
  isOpen: boolean;
  onClose: () => void;
  selectedProducts: Product[];
  onSuccess: () => void;
}

export default function BulkMoveStorageModal({
  isOpen,
  onClose,
  selectedProducts,
  onSuccess
}: BulkMoveStorageModalProps) {
  const { t } = useLanguage();
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);
  const [storageLocations, setStorageLocations] = useState<StoredIn[]>([]);
  const [targetStorageId, setTargetStorageId] = useState<string>('');
  const [moveItems, setMoveItems] = useState<MoveItem[]>([]);

  const loadStorage = useCallback(async () => {
    try {
      const data = await adminFetchStorage();
      setStorageLocations(data);
    } catch {
      toast.error(t('pages.admin.inventory.move_modal.alerts.load_error', 'Failed to load storage locations'));
    }
  }, [toast, t]);

  const initializeMoveItems = useCallback(() => {
    const items: MoveItem[] = [];
    selectedProducts.forEach(product => {
      if (product.stored_in && product.stored_in.length > 0) {
        product.stored_in.forEach(loc => {
          items.push({
            product,
            sourceLocation: loc,
            currentQuantity: loc.quantity
          });
        });
      }
    });
    setMoveItems(items);
  }, [selectedProducts]);

  useEffect(() => {
    if (isOpen) {
      loadStorage();
      initializeMoveItems();
    }
  }, [isOpen, initializeMoveItems, loadStorage]);

  const handleQuantityChange = (index: number, delta: number) => {
    setMoveItems(prev => prev.map((item, i) => {
      if (i !== index) return item;
      const newQty = Math.max(0, item.currentQuantity + delta);
      // Hard limit: cannot increase above original quantity (because we are moving FROM here)
      if (newQty > item.sourceLocation.quantity) return item;
      return { ...item, currentQuantity: newQty };
    }));
  };

  const handleSubmit = async () => {
    if (!targetStorageId) {
      toast.error(t('pages.admin.inventory.move_modal.alerts.select_target', 'Please select a target storage location'));
      return;
    }

    const moves = moveItems
      .filter(item => item.sourceLocation.quantity - item.currentQuantity > 0)
      .map(item => ({
        product_id: item.product.id,
        from_storage_id: item.sourceLocation.stored_in_id,
        quantity: item.sourceLocation.quantity - item.currentQuantity
      }));

    if (moves.length === 0) {
      toast.error(t('pages.admin.inventory.move_modal.alerts.no_items', 'No items to move'));
      return;
    }

    setLoading(true);
    try {
      await adminBulkMoveStorage({
        target_storage_id: targetStorageId,
        moves
      });
      toast.success(t('pages.admin.inventory.move_modal.alerts.success', 'Products relocated successfully'));
      onSuccess();
      onClose();
    } catch (err) {
      const message = err instanceof Error ? err.message : t('pages.admin.inventory.move_modal.alerts.error', 'Failed to relocate products');
      toast.error(message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={t('pages.admin.inventory.move_modal.title', 'Bulk Relocation')}
      maxWidth="max-w-6xl"
    >
      <div className="p-6 space-y-6">
        {/* Global Target Selection */}
        <div className="bg-bg-header/30 p-4 rounded-lg border border-border-main/50 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <span className="text-sm font-medium uppercase tracking-wider text-text-muted">{t('pages.admin.inventory.move_modal.target_label', 'Target Storage:')}</span>
            <select
              value={targetStorageId}
              onChange={(e) => setTargetStorageId(e.target.value)}
              className="bg-bg-surface border border-border-main rounded px-3 py-2 text-text-main focus:outline-none focus:border-accent-primary"
            >
              <option value="">{t('pages.admin.inventory.move_modal.target_placeholder', 'Select Target...')}</option>
              {storageLocations.map(loc => (
                <option key={loc.id} value={loc.id}>{loc.name}</option>
              ))}
            </select>
          </div>
          <div className="text-sm text-text-muted italic">
            {t('pages.admin.inventory.move_modal.target_desc', 'All items will be moved to this location')}
          </div>
        </div>

        {/* Move Table */}
        <div className="border border-border-main rounded-lg overflow-hidden">
          <table className="w-full text-left border-collapse">
            <thead className="bg-bg-header text-text-on-header uppercase text-xs tracking-widest font-bold">
              <tr>
                <th className="p-3">{t('pages.admin.inventory.move_modal.table.image', 'Image')}</th>
                <th className="p-3">{t('pages.admin.inventory.move_modal.table.product_name', 'Product Name')}</th>
                <th className="p-3">{t('pages.admin.inventory.move_modal.table.set', 'Set')}</th>
                <th className="p-3">{t('pages.admin.inventory.move_modal.table.code', 'Code')}</th>
                <th className="p-3">{t('pages.admin.inventory.move_modal.table.foil', 'Foil')}</th>
                <th className="p-3">{t('pages.admin.inventory.move_modal.table.variant', 'Variant')}</th>
                <th className="p-3">{t('pages.admin.inventory.move_modal.table.source', 'Source')}</th>
                <th className="p-3">{t('pages.admin.inventory.move_modal.table.stock', 'Stock')}</th>
                <th className="p-3 text-center">{t('pages.admin.inventory.move_modal.table.to_move', 'To Move')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-main/30">
              {moveItems.map((item, idx) => {
                const toMove = item.sourceLocation.quantity - item.currentQuantity;
                return (
                  <tr key={`${item.product.id}-${item.sourceLocation.stored_in_id}`} className="hover:bg-bg-header/10 transition-colors">
                    <td className="p-3">
                      <div className="w-12 h-16 bg-bg-surface border border-border-main rounded overflow-hidden flex items-center justify-center">
                        {item.product.image_url ? (
                          <div className="relative w-full h-full">
                            <Image 
                              src={item.product.image_url} 
                              alt={item.product.name}
                              fill
                              className="object-cover"
                              unoptimized={item.product.tcg === 'mtg'}
                            />
                          </div>
                        ) : (
                          <span className="text-[10px] text-text-muted">{t('pages.admin.inventory.move_modal.no_image', 'No Image')}</span>
                        )}
                      </div>
                    </td>
                    <td className="p-3 max-w-[200px]">
                      <div className="font-medium text-text-main truncate" title={item.product.name}>{item.product.name}</div>
                      <div className="text-[10px] text-text-muted uppercase tracking-tighter">{item.product.tcg}</div>
                    </td>
                    <td className="p-3 text-sm text-text-muted">
                      {item.product.set_name || '-'}
                    </td>
                    <td className="p-3 text-sm font-mono text-text-muted">
                      {item.product.set_code || '-'}
                    </td>
                    <td className="p-3">
                      <span className={`text-[10px] px-2 py-0.5 rounded-full uppercase font-bold ${
                        item.product.foil_treatment === 'non_foil' ? 'bg-bg-surface text-text-muted border border-border-main' : 'bg-accent-primary/20 text-accent-primary border border-accent-primary/30'
                      }`}>
                        {resolveLabel(item.product.foil_treatment, FOIL_LABELS)}
                      </span>
                    </td>
                    <td className="p-3 text-sm text-text-muted">
                      {resolveLabel(item.product.card_treatment, TREATMENT_LABELS) || t('pages.admin.inventory.move_modal.regular_variant', 'Regular')}
                    </td>
                    <td className="p-3 text-sm font-medium text-accent-header">
                      {item.sourceLocation.name}
                    </td>
                    <td className="p-3">
                      <div className="flex items-center gap-3">
                        <button
                          onClick={() => handleQuantityChange(idx, -1)}
                          className="w-7 h-7 flex items-center justify-center rounded border border-border-main hover:border-accent-primary hover:text-accent-primary transition-colors disabled:opacity-30"
                          disabled={item.currentQuantity <= 0}
                        >
                          -
                        </button>
                        <span className="w-8 text-center font-mono text-lg font-bold">
                          {item.currentQuantity}
                        </span>
                        <button
                          onClick={() => handleQuantityChange(idx, 1)}
                          className="w-7 h-7 flex items-center justify-center rounded border border-border-main hover:border-accent-primary hover:text-accent-primary transition-colors disabled:opacity-30"
                          disabled={item.currentQuantity >= item.sourceLocation.quantity}
                        >
                          +
                        </button>
                      </div>
                    </td>
                    <td className="p-3 text-center">
                      <div className={`text-xl font-bold ${toMove > 0 ? 'text-accent-primary animate-in fade-in' : 'text-text-muted/30'}`}>
                        {toMove}
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>

        {/* Footer Actions */}
        <div className="flex justify-end gap-4 p-2">
          <button
            onClick={onClose}
            className="px-6 py-2 rounded-lg border border-border-main text-text-muted hover:bg-bg-header/50 transition-colors cursor-pointer"
          >
            {t('pages.common.labels.cancel', 'Cancel')}
          </button>
          <button
            onClick={handleSubmit}
            disabled={loading || !targetStorageId || moveItems.every(i => i.currentQuantity === i.sourceLocation.quantity)}
            className="px-10 py-2 rounded-lg bg-accent-primary text-white font-bold hover:bg-accent-primary-hover transition-all shadow-lg shadow-accent-primary/20 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer flex items-center gap-2"
          >
            {loading ? (
              <>
                <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
                {t('pages.admin.inventory.move_modal.processing', 'Processing...')}
              </>
            ) : (
              t('pages.admin.inventory.move_modal.confirm_btn', 'Confirm Relocation')
            )}
          </button>
        </div>
      </div>
    </Modal>
  );
}
