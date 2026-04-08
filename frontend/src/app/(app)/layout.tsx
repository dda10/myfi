/**
 * App layout group — authenticated pages.
 * AppShell (sidebar + header) is already provided by AuthenticatedLayout in root layout.
 * This layout is a pass-through.
 */
export default function AppLayout({ children }: { children: React.ReactNode }) {
  return <>{children}</>;
}
