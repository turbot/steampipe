const lightTheme = (colors) => ({
  color: colors,
  backgroundColor: "rgba(255,255,255,0)",
  textStyle: {},
  title: {
    textStyle: {
      color: "#181717",
    },
    subtextStyle: {
      color: "#999999",
    },
  },
  line: {
    itemStyle: {
      borderWidth: "2",
    },
    lineStyle: {
      width: "3",
    },
    symbolSize: "8",
    symbol: "emptyCircle",
    smooth: false,
  },
  radar: {
    itemStyle: {
      borderWidth: "2",
    },
    lineStyle: {
      width: "3",
    },
    symbolSize: "8",
    symbol: "emptyCircle",
    smooth: false,
  },
  bar: {
    itemStyle: {
      barBorderWidth: 0,
      barBorderColor: "#ccc",
    },
  },
  pie: {
    itemStyle: {
      borderWidth: 0,
      borderColor: "#ccc",
    },
  },
  scatter: {
    itemStyle: {
      borderWidth: 0,
      borderColor: "#ccc",
    },
  },
  boxplot: {
    itemStyle: {
      borderWidth: 0,
      borderColor: "#ccc",
    },
  },
  parallel: {
    itemStyle: {
      borderWidth: 0,
      borderColor: "#ccc",
    },
  },
  sankey: {
    itemStyle: {
      borderWidth: 0,
      borderColor: "#ccc",
    },
  },
  funnel: {
    itemStyle: {
      borderWidth: 0,
      borderColor: "#ccc",
    },
  },
  gauge: {
    itemStyle: {
      borderWidth: 0,
      borderColor: "#ccc",
    },
  },
  candlestick: {
    itemStyle: {
      color: "#e6a0d2",
      color0: "transparent",
      borderColor: "#e6a0d2",
      borderColor0: "#3fb1e3",
      borderWidth: "2",
    },
  },
  graph: {
    itemStyle: {
      borderWidth: 0,
      borderColor: "#ccc",
    },
    lineStyle: {
      width: "1",
      color: "#cccccc",
    },
    symbolSize: "8",
    symbol: "emptyCircle",
    smooth: false,
    color: ["#3fb1e3", "#6be6c1", "#626c91", "#a0a7e6", "#c4ebad", "#96dee8"],
    label: {
      color: "#ffffff",
    },
  },
  map: {
    itemStyle: {
      normal: {
        areaColor: "#eeeeee",
        borderColor: "#aaaaaa",
        borderWidth: 0.5,
      },
      emphasis: {
        areaColor: "rgba(63,177,227,0.25)",
        borderColor: "#3fb1e3",
        borderWidth: 1,
      },
    },
    label: {
      normal: {
        textStyle: {
          color: "#ffffff",
        },
      },
      emphasis: {
        textStyle: {
          color: "#3fb1e3",
        },
      },
    },
  },
  geo: {
    itemStyle: {
      normal: {
        areaColor: "#eeeeee",
        borderColor: "#aaaaaa",
        borderWidth: 0.5,
      },
      emphasis: {
        areaColor: "rgba(63,177,227,0.25)",
        borderColor: "#3fb1e3",
        borderWidth: 1,
      },
    },
    label: {
      normal: {
        textStyle: {
          color: "#ffffff",
        },
      },
      emphasis: {
        textStyle: {
          color: "#3fb1e3",
        },
      },
    },
  },
  categoryAxis: {
    axisLine: {
      show: true,
      lineStyle: {
        color: "#cccccc",
      },
    },
    axisTick: {
      show: false,
      lineStyle: {
        color: "#333",
      },
    },
    axisLabel: {
      show: true,
      textStyle: {
        color: "#999999",
      },
    },
    splitLine: {
      show: true,
      lineStyle: {
        color: ["#eeeeee"],
      },
    },
    splitArea: {
      show: false,
      areaStyle: {
        color: ["rgba(250,250,250,0.05)", "rgba(200,200,200,0.02)"],
      },
    },
  },
  valueAxis: {
    axisLine: {
      show: true,
      lineStyle: {
        color: "#cccccc",
      },
    },
    axisTick: {
      show: false,
      lineStyle: {
        color: "#333",
      },
    },
    axisLabel: {
      show: true,
      textStyle: {
        color: "#999999",
      },
    },
    splitLine: {
      show: true,
      lineStyle: {
        color: ["#eeeeee"],
      },
    },
    splitArea: {
      show: false,
      areaStyle: {
        color: ["rgba(250,250,250,0.05)", "rgba(200,200,200,0.02)"],
      },
    },
  },
  logAxis: {
    axisLine: {
      show: true,
      lineStyle: {
        color: "#cccccc",
      },
    },
    axisTick: {
      show: false,
      lineStyle: {
        color: "#333",
      },
    },
    axisLabel: {
      show: true,
      textStyle: {
        color: "#999999",
      },
    },
    splitLine: {
      show: true,
      lineStyle: {
        color: ["#eeeeee"],
      },
    },
    splitArea: {
      show: false,
      areaStyle: {
        color: ["rgba(250,250,250,0.05)", "rgba(200,200,200,0.02)"],
      },
    },
  },
  timeAxis: {
    axisLine: {
      show: true,
      lineStyle: {
        color: "#cccccc",
      },
    },
    axisTick: {
      show: false,
      lineStyle: {
        color: "#333",
      },
    },
    axisLabel: {
      show: true,
      textStyle: {
        color: "#999999",
      },
    },
    splitLine: {
      show: true,
      lineStyle: {
        color: ["#eeeeee"],
      },
    },
    splitArea: {
      show: false,
      areaStyle: {
        color: ["rgba(250,250,250,0.05)", "rgba(200,200,200,0.02)"],
      },
    },
  },
  toolbox: {
    iconStyle: {
      normal: {
        borderColor: "#999999",
      },
      emphasis: {
        borderColor: "#666666",
      },
    },
  },
  legend: {
    textStyle: {
      color: "#999999",
    },
  },
  tooltip: {
    axisPointer: {
      lineStyle: {
        color: "#cccccc",
        width: 1,
      },
      crossStyle: {
        color: "#cccccc",
        width: 1,
      },
    },
  },
  timeline: {
    lineStyle: {
      color: "#626c91",
      width: 1,
    },
    itemStyle: {
      normal: {
        color: "#626c91",
        borderWidth: 1,
      },
      emphasis: {
        color: "#626c91",
      },
    },
    controlStyle: {
      normal: {
        color: "#626c91",
        borderColor: "#626c91",
        borderWidth: 0.5,
      },
      emphasis: {
        color: "#626c91",
        borderColor: "#626c91",
        borderWidth: 0.5,
      },
    },
    checkpointStyle: {
      color: "#3fb1e3",
      borderColor: "#3fb1e3",
    },
    label: {
      normal: {
        textStyle: {
          color: "#626c91",
        },
      },
      emphasis: {
        textStyle: {
          color: "#626c91",
        },
      },
    },
  },
  visualMap: {
    color: ["#2a99c9", "#afe8ff"],
  },
  dataZoom: {
    backgroundColor: "rgba(255,255,255,0)",
    dataBackgroundColor: "rgba(222,222,222,1)",
    fillerColor: "rgba(114,230,212,0.25)",
    handleColor: "#cccccc",
    handleSize: "100%",
    textStyle: {
      color: "#999999",
    },
  },
  markPoint: {
    label: {
      color: "#ffffff",
    },
    emphasis: {
      label: {
        color: "#ffffff",
      },
    },
  },
});

export default lightTheme;
