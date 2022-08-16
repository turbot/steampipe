import Properties from "./Properties";
import Tooltip from "./Tooltip";
import { circleGetBezierPath, getEdgeParams } from "./utils";
import { EdgeText, useStore } from "react-flow-renderer";
import { useCallback } from "react";

const FloatingEdge = ({
  id,
  source,
  target,
  markerEnd,
  style,
  labelStyle,
  labelBgStyle,
  labelShowBg,
  labelBgPadding,
  labelBgBorderRadius,
  data: { row_data, label },
}) => {
  const sourceNode = useStore(
    useCallback((store) => store.nodeInternals.get(source), [source])
  );
  const targetNode = useStore(
    useCallback((store) => store.nodeInternals.get(target), [target])
  );

  if (!sourceNode || !targetNode) {
    return null;
  }

  const { sx, sy, mx, my, tx, ty, sourcePos, targetPos } = getEdgeParams(
    sourceNode,
    targetNode
  );

  //const d = getSimpleBezierPath({
  const d = circleGetBezierPath({
    sourceX: sx,
    sourceY: sy,
    sourcePosition: sourcePos,
    targetPosition: targetPos,
    targetX: tx,
    targetY: ty,
  });

  const edge = (
    <g className="relative">
      <path
        id={id}
        // className="stroke-[0.5] stroke-[#aaa]"
        className="react-flow__edge-path"
        d={d}
        markerEnd={markerEnd}
        style={style}
      />
      <foreignObject
        className="flex items-center text-xs bg-dashboard-panel text-foreground-light italic z-50"
        height="16"
        width="60"
        x={mx - 30}
        y={my - 8}
        requiredExtensions="http://www.w3.org/1999/xhtml"
      >
        <span className="block items-center truncate" title={label}>
          {label}
        </span>
      </foreignObject>
      {/*<EdgeText*/}
      {/*  // className="italic text-[5px] fill-foreground-light"*/}
      {/*  // className="react-flow__edge-text"*/}
      {/*  // className="text-alert"*/}
      {/*  x={mx}*/}
      {/*  y={my}*/}
      {/*  label={label}*/}
      {/*  labelStyle={labelStyle}*/}
      {/*  labelShowBg={labelShowBg}*/}
      {/*  labelBgStyle={labelBgStyle}*/}
      {/*  labelBgPadding={labelBgPadding}*/}
      {/*  labelBgBorderRadius={labelBgBorderRadius}*/}
      {/*/>*/}
    </g>
  );

  return (
    <>
      {row_data && row_data.properties && (
        <Tooltip
          overlay={<Properties properties={row_data.properties} />}
          title={label}
        >
          {edge}
        </Tooltip>
      )}
      {(!row_data || !row_data.properties) && edge}
    </>
  );
};

export default FloatingEdge;
