import type { ReactNode } from 'react';
import TopBar from './TopBar';

interface LayoutProps {
  children: ReactNode;
}

const Layout = ({ children }: LayoutProps) => {
  return (
    <div className="min-h-screen bg-gray-50">
      <TopBar />
      <main>{children}</main>
    </div>
  );
};

export default Layout;