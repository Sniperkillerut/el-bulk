import Skeleton from '../ui/Skeleton';

interface ProductGridSkeletonProps {
  count?: number;
}

export default function ProductGridSkeleton({ count = 12 }: ProductGridSkeletonProps) {
  return (
    <div className="product-grid-skeleton grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="flex flex-col gap-3 p-3 bg-kraft-mid opacity-50 border-2 border-kraft-dark rounded-sm">
          {/* Card Image Placeholder */}
          <Skeleton 
             className="aspect-[63/88] rounded-sm" 
          />
          
          {/* Badge List Placeholder */}
          <div className="flex gap-1">
            <Skeleton width="40px" height="12px" />
            <Skeleton width="60px" height="12px" />
          </div>

          {/* Title Placeholder */}
          <Skeleton width="100%" height="20px" />
          
          {/* Subtitle Placeholder */}
          <Skeleton width="70%" height="14px" />

          {/* Price/CTA Placeholder */}
          <div className="flex items-center justify-between mt-2 pt-2 border-t border-kraft-dark">
            <Skeleton width="60px" height="24px" />
            <Skeleton width="50px" height="28px" />
          </div>
        </div>
      ))}
    </div>
  );
}
