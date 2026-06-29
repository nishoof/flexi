/** Shown while checking whether the user has a valid session */
export default function LoadingView() {
  return (
    <div className="flex flex-col items-center justify-center gap-4 py-24">
      <div
        className="size-8 rounded-full border-2 border-(--border) border-t-(--accent) animate-spin"
        aria-hidden="true"
      />
      <p className="text-sm text-(--foreground)/70">Loading your dashboard...</p>
    </div>
  );
}
