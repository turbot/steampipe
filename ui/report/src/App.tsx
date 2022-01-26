import Report from "./components/reports/layout/Report";
import ReportErrorModal from "./components/reports/ReportErrorModal";
import ReportHeader from "./components/ReportHeader";
import ReportList from "./components/ReportList";
import ReportSelector from "./components/ReportSelector";
import { BreakpointProvider } from "./hooks/useBreakpoint";
import { ReportProvider } from "./hooks/useReport";
import { Route, Routes } from "react-router-dom";

const ReportApp = () => (
  <ReportProvider>
    <ReportHeader>
      <ReportSelector />
    </ReportHeader>
    <ReportErrorModal />
    <ReportList />
    <Report />
  </ReportProvider>
);

const App = () => (
  <BreakpointProvider>
    <Routes>
      <Route path="/" element={<ReportApp />} />
      <Route path="/:reportName" element={<ReportApp />} />
    </Routes>
  </BreakpointProvider>
);

export default App;

/**
 * steampipe report <command>
 *
 * 1) Socket and Static file server started - socket server is aware of context/workspace it's launched in
 *    e.g. what workspace and/or specific report path are we interested in?
 * 2) Browser opened at Static file server endpoint e.g. http://localhost:3000
 * 3) Browser connects to the socket server
 * 4) Socket server broadcasts its context to clients on connect
 * 5) Browser will display its current context
 * 6) Socket server will broadcast updates to the connected client(s)
 *    TODO handle multiple clients with diff contexts?
 * 7) Browser will handle updates and handle state of report and its panels
 *
 * **/
