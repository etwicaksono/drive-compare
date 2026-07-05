import { NavLink } from 'react-router-dom';

function Sidebar(): React.JSX.Element {
  return (
    <aside className="flex h-screen w-56 flex-col border-r border-neutral-700 bg-neutral-900">
      <div className="px-4 py-5">
        <h1 className="text-sm font-bold tracking-wide text-neutral-300 uppercase">
          PDI
        </h1>
      </div>
      <nav className="flex flex-1 flex-col gap-1 px-2">
        <SidebarLink to="/" label="Compare" />
        <SidebarLink to="/duplicates" label="Duplicates" />
        <SidebarLink to="/settings" label="Settings" />
      </nav>
    </aside>
  );
}

function SidebarLink({ to, label }: { to: string; label: string }): React.JSX.Element {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        `rounded-md px-3 py-2 text-sm transition-colors ${
          isActive
            ? 'bg-neutral-700 text-white font-medium'
            : 'text-neutral-400 hover:bg-neutral-800 hover:text-white'
        }`
      }
    >
      {label}
    </NavLink>
  );
}

export default Sidebar;
