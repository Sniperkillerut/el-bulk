import ProductGridSkeleton from '@/components/skeletons/ProductGridSkeleton';

export default function SinglesLoading() {
  return (
    <div className="min-h-screen pb-20">
      {/* Skeleton for Header Section */}
      <section className="bg-kraft-mid border-b-4 border-kraft-dark py-10 md:py-12 px-4 relative overflow-hidden box-lid">
        <div className="centered-container relative z-10 text-center px-4 animate-pulse">
          <div className="h-6 w-32 bg-kraft-dark/20 mx-auto mb-4 rounded-sm" />
          <div className="h-16 w-64 bg-kraft-dark/30 mx-auto mb-4 rounded-sm" />
          <div className="h-4 w-96 bg-kraft-dark/20 mx-auto rounded-sm" />
        </div>
      </section>

      <div className="centered-container px-4 mt-12">
        {/* Skeleton for TCG Selection Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8 mb-20">
            {[...Array(3)].map((_, i) => (
                <div key={i} className="aspect-[16/10] bg-kraft-mid border-2 border-kraft-dark rounded-xl opacity-40 animate-pulse" />
            ))}
        </div>

        {/* Featured Section Skeleton */}
        <section>
          <div className="flex items-center gap-4 mb-8">
            <div className="h-10 w-48 bg-kraft-dark/10 rounded-sm" />
            <div className="h-[2px] w-full bg-kraft-dark/20" />
          </div>

          <ProductGridSkeleton count={12} />
        </section>
      </div>
    </div>
  );
}
