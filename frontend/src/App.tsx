import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import ComparePage from './pages/ComparePage';
import SettingsPage from './pages/SettingsPage';

function App(): React.JSX.Element {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<ComparePage />} />
          <Route path="/settings" element={<SettingsPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
