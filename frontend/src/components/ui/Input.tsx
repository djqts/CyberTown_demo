import { cn } from "@/lib/utils";
import { type InputHTMLAttributes, forwardRef } from "react";

const Input = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(
  ({ className, ...props }, ref) => (
    <input
      ref={ref}
      className={cn(
        "flex h-9 w-full rounded-md border border-[#3B4B5C] bg-[#18202A] px-3 py-1 text-sm text-[#F3F5F7]",
        "placeholder:text-[#B8C2CC]",
        "focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[#7CCB8A]",
        className
      )}
      {...props}
    />
  )
);
Input.displayName = "Input";

export { Input };
