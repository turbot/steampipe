import React, { Fragment, useState } from "react";
import { CloseIcon } from "../../constants/icons";
import { Dialog, Transition } from "@headlessui/react";
import { ModalThemeWrapper, ThemeProvider } from "../../hooks/useTheme";
import { classNames } from "../../utils/styles";

interface ModalProps {
  actions: JSX.Element[];
  allowClickAway?: boolean;
  children: null | string | JSX.Element | JSX.Element[];
  icon?: JSX.Element;
  onClose: () => Promise<void>;
  title: string;
}

const Modal = ({
  actions,
  allowClickAway = false,
  icon,
  children,
  onClose,
  title,
}: ModalProps) => {
  const [open, setOpen] = useState(true);

  const handleOnClose = () => {
    setOpen(false);
    onClose && onClose();
  };

  return (
    <ThemeProvider>
      <Transition.Root show={open} as={Fragment}>
        <Dialog
          as="div"
          auto-reopen="true"
          className="fixed z-10 inset-0 overflow-y-auto"
          onClose={allowClickAway ? onClose : () => {}}
        >
          <ModalThemeWrapper>
            <div className="min-h-screen pt-4 px-4 text-center">
              <Transition.Child
                as={Fragment}
                enter="ease-out duration-300"
                enterFrom="opacity-0"
                enterTo="opacity-100"
                leave="ease-in duration-200"
                leaveFrom="opacity-100"
                leaveTo="opacity-0"
              >
                <Dialog.Overlay className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
              </Transition.Child>

              {/* This element is to trick the browser into centering the modal contents. */}
              <span
                className="inline-block align-middle h-screen"
                aria-hidden="true"
              >
                &#8203;
              </span>
              <Transition.Child
                as={Fragment}
                enter="ease-out duration-300"
                enterFrom="opacity-0 translate-y-0 scale-95"
                enterTo="opacity-100 translate-y-0 scale-100"
                leave="ease-in duration-200"
                leaveFrom="opacity-100 translate-y-0 scale-100"
                leaveTo="opacity-0 translate-y-0 scale-95"
              >
                <div className="inline-block w-full sm:max-w-xl lg:max-w-3xl h-full sm:h-auto align-middle bg-dashboard rounded-lg p-4 text-left overflow-hidden shadow-xl transform transition-all my-8 space-y-4">
                  <div className="flex items-center space-x-3">
                    <div className="flex-shrink-0 flex items-start">{icon}</div>
                    <div className="text-left">
                      <Dialog.Title
                        as="h2"
                        className="text-xl leading-6 font-medium text-foreground"
                      >
                        {title}
                      </Dialog.Title>
                    </div>
                    <div className="absolute top-0 right-0 pt-4 pr-4">
                      <button
                        type="button"
                        className="bg-dashboard rounded-md text-foreground-light hover:text-foreground focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                        onClick={handleOnClose}
                      >
                        <span className="sr-only">Close</span>
                        <CloseIcon className="h-6 w-6" aria-hidden="true" />
                      </button>
                    </div>
                  </div>
                  <div>{children}</div>
                  <div
                    className={classNames(
                      "flex pt-4 border-t border-divide space-x-2 justify-end"
                    )}
                  >
                    <div className="flex flex-1 space-x-3 justify-end">
                      {actions}
                    </div>
                  </div>
                </div>
              </Transition.Child>
            </div>
          </ModalThemeWrapper>
        </Dialog>
      </Transition.Root>
    </ThemeProvider>
  );
};

export default Modal;
