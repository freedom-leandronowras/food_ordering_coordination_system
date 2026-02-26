import { ReactNode } from "react";

type AppShellProps = {
  header: ReactNode;
  sidebar: ReactNode;
  main: ReactNode;
  aside: ReactNode;
  mobileNav: ReactNode;
};

export function AppShell({ header, sidebar, main, aside, mobileNav }: AppShellProps) {
  return (
    <main className="min-h-screen bg-[#f3f7f5] px-4 pb-24 pt-4 text-[#123830] lg:px-8 lg:pb-8">
      <div className="mx-auto flex w-full max-w-[1240px] flex-col gap-5">
        {header}

        <section className="space-y-5 lg:grid lg:grid-cols-[170px_minmax(0,1fr)_320px] lg:gap-5 lg:space-y-0">
          <aside className="hidden lg:block">{sidebar}</aside>
          <section>{main}</section>
          <aside>{aside}</aside>
        </section>

        <section className="lg:hidden">{sidebar}</section>
      </div>

      <nav className="fixed inset-x-4 bottom-4 z-30 rounded-2xl border border-[#dce9e5] bg-white p-2 shadow-lg lg:hidden">
        {mobileNav}
      </nav>
    </main>
  );
}
