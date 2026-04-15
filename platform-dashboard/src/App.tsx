import { BrowserRouter as Router, Navigate, Routes, Route } from 'react-router-dom'
import { ThemeProvider } from './context/ThemeContext'
import { LocalizationProvider } from './context/LocalizationContext'
import Sidebar from './components/layout/Sidebar'
import Overview from './pages/Overview'
import Workers from './pages/Workers'
import Zones from './pages/Zones'
import Analytics from './pages/Analytics'
import Disruptions from './pages/Disruptions'
import ReconciliationOps from './pages/ReconciliationOps'
import GodModeLayout from './pages/god-mode/GodModeLayout'
import { GodModeProvider } from './pages/god-mode/state'

export default function App() {
  return (
    <ThemeProvider>
      <LocalizationProvider>
        <Router
          future={{
            v7_startTransition: true,
            v7_relativeSplatPath: true,
          }}
        >
          <Sidebar>
            <Routes>
              <Route path="/" element={<Overview />} />
              <Route path="/workers" element={<Workers />} />
              <Route path="/zones" element={<Zones />} />
              <Route path="/analytics" element={<Analytics />} />
              <Route path="/reconciliation" element={<ReconciliationOps />} />
              <Route path="/disruptions" element={<Disruptions />} />
              <Route
                path="/batches"
                element={(
                  <GodModeProvider>
                    <GodModeLayout />
                  </GodModeProvider>
                )}
              />
              <Route
                path="/god-mode"
                element={(
                  <GodModeProvider>
                    <GodModeLayout />
                  </GodModeProvider>
                )}
              />
              <Route path="/god-mode/batch-simulation" element={<Navigate to="/batches" replace />} />
              <Route path="/weekly-cycle" element={<Navigate to="/" replace />} />
              <Route path="/payout-ops" element={<Navigate to="/" replace />} />
              <Route path="/synthetic-data" element={<Navigate to="/" replace />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </Sidebar>
        </Router>
      </LocalizationProvider>
    </ThemeProvider>
  )
}
