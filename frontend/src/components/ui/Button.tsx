import { cn } from "@/lib/utils";
import { type ButtonHTMLAttributes, forwardRef } from "react";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "default" | "ghost" | "outline";
  size?: "sm" | "default";
}

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "default", size = "default", ...props }, ref) => (
    <button
      ref={ref}
      className={cn(
        "inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors",
        "focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[#7CCB8A]",
        "disabled:pointer-events-none disabled:opacity-50",
        variant === "ghost" && "hover:bg-white/10 text-[#B8C2CC] hover:text-[#F3F5F7]",
        variant === "outline" && "border border-[#3B4B5C] bg-transparent hover:bg-white/5 text-[#F3F5F7]",
        variant === "default" && "bg-[#243241] hover:bg-[#3B4B5C] text-[#F3F5F7]",
        size === "sm" && "h-8 px-2 text-xs",
        size === "default" && "h-9 px-3",
        className
      )}
      {...props}
    />
  )
);
Button.displayName = "Button";

export { Button };
