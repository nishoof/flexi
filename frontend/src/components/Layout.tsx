function Layout({ children }: { children: React.ReactNode }) {
  return (
    <main className="max-w-5xl p-8 mx-auto">
      {children}
    </main>
  );
}

export default Layout;
