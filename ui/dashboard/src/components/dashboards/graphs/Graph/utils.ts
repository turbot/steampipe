import { Position } from "reactflow";

export const CIRCLE_SIZE = 50;

// this helper function returns the intersection point
// of the line between the center of the intersectionNode and the target node
function getNodeIntersection(intersectionNode, targetNode) {
  // https://math.stackexchange.com/questions/1724792/an-algorithm-for-finding-the-intersection-point-between-a-center-of-vision-and-a
  const { width: intersectionNodeWidth, position: intersectionNodePosition } =
    intersectionNode;
  const targetPosition = targetNode.position;

  const w = intersectionNodeWidth / 2;
  //const h = intersectionNodeHeight / 2;
  const h = CIRCLE_SIZE / 2;

  const x2 = intersectionNodePosition.x + w;
  const y2 = intersectionNodePosition.y + h;
  const x1 = targetPosition.x + w;
  const y1 = targetPosition.y + h;

  // Algorithm from https://stackoverflow.com/a/18009621.1,
  const padding = 5;
  const r = CIRCLE_SIZE / 2 + padding;
  const phi = Math.atan2(y1 - y2, x1 - x2);
  const x = x2 + r * Math.cos(phi);
  const y = y2 + r * Math.sin(phi);

  return { x, y };
  /*

  // Use linear algebra for slope and intercept: y = mx + b
  const m = (y2 - y1) / (x2 - x1)
  const b = y1 - (m * x1)

  // Now, use pythagoras theorem for x and y difference. We have a circle out
  // from (x2, y2) with radius r. So, the intercept point (xi, yi) must be
  // subject to a delta from x2 to xi of xd and from y2 to yi of yd, and:
  // r^2 = xd^2 + yd^2.
  //
  // But we also know yd = m*xd + b
  //
  // So:
  //   r^2 = xd^2 + (m*xd + b)^2
  //
  //
  // x =
  //

  const gap = Math.Sqrt(50^2 / 2)

  let x = x2
  let y = y2
  if (x1 <
  return { x: x2, y: y2 };
  */

  /*

  const xx1 = (x1 - x2) / (2 * w) - (y1 - y2) / (2 * h);
  const yy1 = (x1 - x2) / (2 * w) + (y1 - y2) / (2 * h);
  const a = 1 / (Math.abs(xx1) + Math.abs(yy1));
  const xx3 = a * xx1;
  const yy3 = a * yy1;
  const x = w * (xx3 + yy3) + x2;
  const y = h * (-xx3 + yy3) + y2;

  return { x, y }
  */
}

// returns the position (top,right,bottom or right) passed node compared to the intersection point
function getEdgePosition(node, intersectionPoint) {
  const n = { ...node.position, ...node };
  const nx = Math.round(n.x);
  const ny = Math.round(n.y);
  const px = Math.round(intersectionPoint.x);
  const py = Math.round(intersectionPoint.y);

  if (px <= nx + 1) {
    return Position.Left;
  }
  if (px >= nx + n.width - 1) {
    return Position.Right;
  }
  if (py <= ny + 1) {
    return Position.Top;
  }
  if (py >= n.y + n.height - 1) {
    return Position.Bottom;
  }

  return Position.Top;
}

// returns the parameters (sx, sy, tx, ty, sourcePos, targetPos) you need to create an edge
export function getEdgeParams(source, target) {
  const sourceIntersectionPoint = getNodeIntersection(source, target);
  const targetIntersectionPoint = getNodeIntersection(target, source);

  const sourcePos = getEdgePosition(source, sourceIntersectionPoint);
  const targetPos = getEdgePosition(target, targetIntersectionPoint);

  const mx =
    sourceIntersectionPoint.x +
    (targetIntersectionPoint.x - sourceIntersectionPoint.x) / 2;
  const my =
    sourceIntersectionPoint.y +
    (targetIntersectionPoint.y - sourceIntersectionPoint.y) / 2;

  return {
    sx: sourceIntersectionPoint.x,
    sy: sourceIntersectionPoint.y,
    tx: targetIntersectionPoint.x,
    ty: targetIntersectionPoint.y,
    mx,
    my,
    sourcePos,
    targetPos,
  };
}

/*
export function createNodesAndEdges() {
  const nodes = [];
  const edges = [];
  const center = { x: window.innerWidth / 2, y: window.innerHeight / 2 };

  nodes.push({ id: "target", data: { label: "Target" }, position: center });

  for (let i = 0; i < 8; i++) {
    const degrees = i * (360 / 8);
    const radians = degrees * (Math.PI / 180);
    const x = 250 * Math.cos(radians) + center.x;
    const y = 250 * Math.sin(radians) + center.y;

    nodes.push({ id: `${i}`, data: { label: "Source" }, position: { x, y } });

    edges.push({
      id: `edge-${i}`,
      target: "target",
      source: `${i}`,
      type: "floating",
      markerEnd: {
        type: MarkerType.Arrow,
      },
    });
  }

  return { nodes, edges };
}
*/

function circleGetControlWithCurvature({ x1, y1, x2, y2, c }) {
  let ctX;
  let ctY;

  if (x1 > x2) {
    ctX = x1 - calculateControlOffset(x2 - x1, c);
  } else {
    ctX = x1 + calculateControlOffset(x1 - x2, c);
  }

  if (y1 > y2) {
    ctY = y1 - calculateControlOffset(y2 - y1, c);
  } else {
    ctY = y1 + calculateControlOffset(y1 - y2, c);
  }

  return [ctX, ctY];
}

function calculateControlOffset(distance, curvature) {
  if (distance >= 0) {
    return 0.5 * distance;
  } else {
    return curvature * 25 * Math.sqrt(-distance);
  }
}

export function circleGetBezierPath({
  sourceX,
  sourceY,
  targetX,
  targetY,
  curvature = 0,
}) {
  const [sourceControlX, sourceControlY] = circleGetControlWithCurvature({
    x1: sourceX,
    y1: sourceY,
    x2: targetX,
    y2: targetY,
    c: curvature,
  });
  const [targetControlX, targetControlY] = circleGetControlWithCurvature({
    x1: targetX,
    y1: targetY,
    x2: sourceX,
    y2: sourceY,
    c: curvature,
  });
  return `M${sourceX},${sourceY} C${sourceControlX},${sourceControlY} ${targetControlX},${targetControlY} ${targetX},${targetY}`;
}

export const buildLabelTextShadow = (color: string) =>
  `3px 0px 0 ${color},2.837451725101904px 0.9740984076140504px 0 ${color},2.3674215281891806px 1.8426381380690033px 0 ${color},1.6408444743672808px 2.5114994347875856px 0 ${color},0.7364564614223977px 2.908200797817991px 0 ${color},-0.24773803641699682px 2.9897534790200098px 0 ${color},-1.2050862739589083px 2.747319979965172px 0 ${color},-2.0318447148772227px 2.207171732019395px 0 ${color},-2.638421253619467px 1.427842179111221px 0 ${color},-2.959083910208167px 0.4937837708422021px 0 ${color},-2.959083910208167px -0.4937837708422014px 0 ${color},-2.6384212536194678px -1.4278421791112192px 0 ${color},-2.031844714877223px -2.2071717320193946px 0 ${color},-1.2050862739589072px -2.747319979965173px 0 ${color},-0.2477380364169982px -2.9897534790200098px 0 ${color},0.7364564614223964px -2.9082007978179916px 0 ${color},1.6408444743672796px -2.511499434787586px 0 ${color},2.3674215281891815px -1.842638138069002px 0 ${color},2.837451725101904px -0.9740984076140512px 0 ${color}`;
