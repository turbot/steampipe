import { Chart } from "chart.js";
import { useCallback, useEffect } from "react";

const usePrintCharts = () => {
  const resizeCharts = useCallback(() => {
    console.log("Resizing charts...");
    for (const id in Chart.instances) {
      // Chart.instances[id].resize(200, 200);
      Chart.instances[id].resize();
    }
  }, []);

  // const restoreCharts = useCallback(() => {
  //   console.log("Restoring charts...");
  //   for (const id in Chart.instances) {
  //     Chart.instances[id].resize();
  //   }
  // }, []);

  useEffect(() => {
    console.log("Adding print listeners");

    if ("matchMedia" in window) {
      window.matchMedia("print").addEventListener("beforeprint", resizeCharts);
    } else {
      window.addEventListener("beforeprint", resizeCharts);
    }

    // window.addEventListener("beforeprint", resizeCharts);
    // window.addEventListener("afterprint", restoreCharts);
    return () => {
      console.log("Removing print listeners");
      window.removeEventListener("beforeprint", resizeCharts);
      // window.removeEventListener("afterprint", restoreCharts);
    };
  }, []);
};

export default usePrintCharts;
