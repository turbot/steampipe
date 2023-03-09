import { PanelDefinition } from "../types";
import { saveAs } from "file-saver";
import { timestampForFilename } from "../utils/date";
import { useCallback, useState } from "react";
import { useDashboard } from "./useDashboard";
import { usePapaParse } from "react-papaparse";

const useDownloadPanelData = (definition: PanelDefinition) => {
  const { selectedDashboard } = useDashboard();
  const { jsonToCSV } = usePapaParse();
  const [processing, setProcessing] = useState(false);

  const downloadQueryData = useCallback(async () => {
    if (!definition.data) {
      return;
    }
    setProcessing(true);
    const data = definition.data;
    const colNames = data.columns.map((c) => c.name);
    let csvRows: any[] = [];

    const jsonbColIndices = data.columns
      .filter((i) => i.data_type === "JSONB")
      .map((i) => data.columns.indexOf(i)); // would return e.g. [3,6,9]

    for (const row of data.rows) {
      // Deep copy the row or else it will update
      // the values in query output
      const csvRow: any[] = [];
      colNames.forEach((col, index) => {
        csvRow[index] = jsonbColIndices.includes(index)
          ? JSON.stringify(row[col])
          : row[col];
      });
      csvRows.push(csvRow);
    }

    const csv = jsonToCSV([colNames, ...csvRows]);
    const blob = new Blob([csv], { type: "text/csv;charset=utf-8" });

    saveAs(
      blob,
      `${(
        selectedDashboard?.full_name ||
        definition.dashboard ||
        ""
      ).replaceAll(".", "_")}_${definition.panel_type}_${timestampForFilename(
        Date.now()
      )}.csv`
    );
    setProcessing(false);
  }, [definition, jsonToCSV, selectedDashboard]);

  return { download: downloadQueryData, processing };
};

export default useDownloadPanelData;
