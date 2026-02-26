import { cn } from "@/lib/utils";

type InlineFeedbackProps = {
  message: string;
  tone?: "success" | "error" | "info";
  className?: string;
};

const toneStyles: Record<NonNullable<InlineFeedbackProps["tone"]>, string> = {
  success: "border-sl-dbe9e4 bg-sl-eef7f4 text-sl-235f56",
  error: "border-sl-f0d1cf bg-sl-fff6f5 text-sl-8f352c",
  info: "border-sl-dce9e5 bg-sl-f8fbfa text-sl-4e6f66",
};

const dotStyles: Record<NonNullable<InlineFeedbackProps["tone"]>, string> = {
  success: "bg-sl-1f6f64 text-sl-ffffff",
  error: "bg-sl-b33e2d text-sl-ffffff",
  info: "bg-sl-607b74 text-sl-ffffff",
};

const dotSymbols: Record<NonNullable<InlineFeedbackProps["tone"]>, string> = {
  success: "✓",
  error: "!",
  info: "i",
};

export function InlineFeedback({ message, tone = "info", className }: InlineFeedbackProps) {
  return (
    <div
      role={tone === "error" ? "alert" : "status"}
      aria-live={tone === "error" ? "assertive" : "polite"}
      className={cn(
        "motion-feedback flex items-start gap-3 rounded-2xl border px-3 py-2 text-sm font-medium shadow-sm",
        toneStyles[tone],
        className,
      )}
    >
      <span className={cn("mt-0.5 inline-flex h-5 w-5 items-center justify-center rounded-full text-xs", dotStyles[tone])}>
        {dotSymbols[tone]}
      </span>
      <p>{message}</p>
    </div>
  );
}
