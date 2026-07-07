import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import ComparePage from './pages/ComparePage';
import DuplicatesPage from './pages/DuplicatesPage';
import ResizeAuditPage from './pages/ResizeAuditPage';
import SettingsPage from './pages/SettingsPage';

function App(): React.JSX.Element {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<ComparePage />} />
          <Route path="/duplicates" element={<DuplicatesPage />} />
          <Route path="/resize-audit" element={<ResizeAuditPage />} />
          <Route path="/settings" element={<SettingsPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
