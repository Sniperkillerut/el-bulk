import ProductGrid from '@/components/ProductGrid';

export default function StoreExclusivesPage() {
  return (
    <ProductGrid
      tcg="all"
      category="store_exclusives"
      titleKey="pages.store_exclusives.title"
      subtitleKey="pages.store_exclusives.subtitle"
      title="STORE EXCLUSIVES"
      subtitle="Custom Commander decks, proxy kits, and other premium items crafted in-house."
    />
  );
}
