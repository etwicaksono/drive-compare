import { Outlet } from 'react-router-dom';
import Sidebar from './Sidebar';

function Layout(): React.JSX.Element {
  return (
    <div className="flex h-screen bg-neutral-950 text-white">
      <Sidebar />
      <main className="flex-1 overflow-auto p-6">
        <Outlet />
      </main>
    </div>
  );
}

export default Layout;
