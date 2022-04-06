import {
  ArrowsExpandIcon as ArrowsExpandIconOutline,
  ClipboardListIcon as ClipboardListIconOutline,
  ExclamationCircleIcon as ExclamationCircleIconOutline,
  MinusSmIcon as MinusSmIconOutline,
  PlusSmIcon as PlusSmIconOutline,
  SearchIcon as SearchIconOutline,
  SaveIcon as SaveIconOutline,
  XIcon as XIconOutline,
} from "@heroicons/react/outline";
import {
  CheckIcon as CheckIconSolid,
  QuestionMarkCircleIcon as QuestionMarkCircleIconSolid,
  ChevronDownIcon as ChevronDownIconSolid,
  ChevronUpIcon as ChevronUpIconSolid,
  ClipboardCheckIcon as ClipboardCheckIconSolid,
  InformationCircleIcon as InformationCircleIconSolid,
  XIcon as XIconSolid,
} from "@heroicons/react/solid";

// General
export const ClearIcon = XIconOutline;
export const CloseIcon = XIconOutline;
export const CopyToClipboardIcon = ClipboardListIconOutline;
export const CopyToClipboardSuccessIcon = ClipboardCheckIconSolid;
export const ErrorIcon = ExclamationCircleIconOutline;
export const SearchIcon = SearchIconOutline;
export const SubmitIcon = SaveIconOutline;
export const ZoomIcon = ArrowsExpandIconOutline;

// Benchmark
export const CollapseBenchmarkIcon = MinusSmIconOutline;
export const ExpandBenchmarkIcon = PlusSmIconOutline;

// Control
export const AlarmIcon = XIconSolid;
export const InfoIcon = InformationCircleIconSolid;
export const OKIcon = CheckIconSolid;
export const UnknownIcon = QuestionMarkCircleIconSolid;

// Table
export const SortAscendingIcon = ChevronUpIconSolid;
export const SortDescendingIcon = ChevronDownIconSolid;
