import { cn } from "@/lib/utils";
import { type ReactNode, useState } from "react";

interface TabsProps {
  tabs: { key: string; label: string }[];
  active?: string;
  onChange?: (key: string) => void;
  children?: ReactNode;
}

function Tabs({ tabs, active, onChange, children }: TabsProps) {
  const [internal, setInternal] = useState(tabs[0]?.key ?? "");
  const current = active ?? internal;
  const setCurrent = onChange ?? setInternal;

  return (
    <div>
      <div className="flex gap-1 border-b border-[#3B4B5C] px-1">
        {tabs.map((t) => (
          <button
            key={t.key}
            onClick={() => setCurrent(t.key)}
            className={cn(
              "px-3 py-2 text-sm font-medium transition-colors border-b-2 -mb-px",
              current === t.key
                ? "border-[#7CCB8A] text-[#F3F5F7]"
                : "border-transparent text-[#B8C2CC] hover:text-[#F3F5F7]"
            )}
          >
            {t.label}
          </button>
        ))}
      </div>
      <div className="mt-3">{children}</div>
    </div>
  );
}

export { Tabs };
