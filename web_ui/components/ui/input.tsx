import * as React from "react";

import { cn } from "@/lib/utils";

const Input = React.forwardRef<HTMLInputElement, React.ComponentProps<"input">>(
  ({ className, type, ...props }, ref) => {
    return (
      <input
        type={type}
        className={cn(
          "flex h-10 w-full rounded-2xl border border-[#d7e6e1] bg-[#f8fbfa] px-3 py-2 text-sm text-[#123830] placeholder:text-[#7a908a] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#1f6f64]",
          className,
        )}
        ref={ref}
        {...props}
      />
    );
  },
);
Input.displayName = "Input";

export { Input };
