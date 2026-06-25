import { cn } from "@/lib/utils";
import type { HTMLAttributes } from "react";

interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: "default" | "accent" | "warning" | "error";
}

function Badge({ className, variant = "default", ...props }: BadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium",
        variant === "default" && "bg-white/10 text-[#B8C2CC]",
        variant === "accent" && "bg-[#7CCB8A]/15 text-[#7CCB8A]",
        variant === "warning" && "bg-[#E6B65C]/15 text-[#E6B65C]",
        variant === "error" && "bg-[#E66A6A]/15 text-[#E66A6A]",
        className
      )}
      {...props}
    />
  );
}

export { Badge };
