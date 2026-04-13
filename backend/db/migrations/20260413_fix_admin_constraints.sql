DO $$
BEGIN
    -- 1. Remove NOT NULL constraint from "password_hash"
    ALTER TABLE "public"."admin" ALTER COLUMN "password_hash" DROP NOT NULL;

    -- 2. Handle the missing column error: "avatar_url" does not exist in the schema provided.
    -- If it were to be added or modified, it must exist first. 
    -- Based on the schema provided, we skip the ALTER for "avatar_url".

    -- 3. Fix the "email" column containing NULL values before setting NOT NULL.
    -- We provide a fallback value for existing NULLs to satisfy the constraint.
    UPDATE "public"."admin"
    SET "email" = 'placeholder_' || "id" || '@example.com'
    WHERE "email" IS NULL;

    -- 4. Ensure "username" does not have NULLs (precautionary)
    UPDATE "public"."admin"
    SET "username" = 'user_' || "id"
    WHERE "username" IS NULL;

    -- 5. Apply NOT NULL constraints
    ALTER TABLE "public"."admin" ALTER COLUMN "email" SET NOT NULL;
    ALTER TABLE "public"."admin" ALTER COLUMN "username" SET NOT NULL;
END $$;

 
 
 
 
 