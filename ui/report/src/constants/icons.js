import { CheckIcon, MoonIcon, SunIcon } from "@heroicons/react/solid";
import {
  faCaretDown as falCaretDown,
  faCaretUp as falCaretUp,
  faChevronDoubleLeft as falChevronDoubleLeft,
  faChevronDoubleRight as falChevronDoubleRight,
  faChevronLeft as falChevronLeft,
  faChevronRight as falChevronRight,
  faCircleNotch as falCircleNotch,
  faExpandArrows as falExpandArrows,
  faHorizontalRule as falHorizontalRule,
  faTag as falTag,
  faTags as falTags,
  faTimes as falTimes,
} from "@fortawesome/pro-light-svg-icons";
import {
  faCheck as fasCheck,
  faExclamationCircle as fasExclamationCircle,
  faInfoCircle as fasInfoCircle,
  faQuestion as fasQuestion,
  faSort as fasSort,
  faTimes as fasTimes,
  faTimesCircle as fasTimesCircle,
} from "@fortawesome/pro-solid-svg-icons";
import { faSteampipe as fabSteampipe } from "../components/Icon/faSteampipe";

// General
export const closeIcon = falTimes;
export const DarkIcon = MoonIcon;
export const emptyIcon = falCircleNotch;
export const errorIcon = fasExclamationCircle;
export const LightIcon = SunIcon;
export const loadingIcon = falCircleNotch;
export const steampipeIcon = fabSteampipe;
export const zoomIcon = falExpandArrows;

// Report primitives
export const openSelectMenuIcon = fasSort;

// Counter
export const alertIcon = fasTimesCircle;
export const nilIcon = falHorizontalRule;
export const infoIcon = fasInfoCircle;

// Control
export const alarmIcon = fasTimes;
export const okIcon = fasCheck;
export const tbdIcon = fasQuestion;

// Resource
export const tagIcon = falTag;
export const tagsIcon = falTags;

// Table
export const firstPageIcon = falChevronDoubleLeft;
export const lastPageIcon = falChevronDoubleRight;
export const nextPageIcon = falChevronRight;
export const previousPageIcon = falChevronLeft;
export const sortAscendingIcon = falCaretUp;
export const sortDescendingIcon = falCaretDown;
