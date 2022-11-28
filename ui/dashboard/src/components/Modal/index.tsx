import { CloseIcon } from "../../constants/icons";
import { Dialog, Transition } from "@headlessui/react";
import { Fragment, useState } from "react";
import { ModalThemeWrapper, ThemeProvider } from "../../hooks/useTheme";

const Modal = ({ icon, message, title }) => {
  const [open, setOpen] = useState(true);

  return (
    <ThemeProvider>
      <Transition.Root show={open} as={Fragment}>
        <Dialog
          as="div"
          static
          className="fixed z-10 inset-0 overflow-y-auto"
          open={open}
          onClose={setOpen}
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
                <div className="inline-block h-full sm:h-auto align-middle bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all my-8 w-full sm:max-w-xl sm:p-6 lg:max-w-3xl">
                  <div className="absolute top-0 right-0 pt-4 pr-4">
                    <button
                      type="button"
                      className="bg-white rounded-md text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                      onClick={() => setOpen(false)}
                    >
                      <span className="sr-only">Close</span>
                      <CloseIcon className="h-6 w-6" aria-hidden="true" />
                    </button>
                  </div>
                  <div className="flex items-start">
                    <div className="flex-shrink-0 flex items-start justify-center h-12 w-12 rounded-full h-12 w-12">
                      {icon}
                    </div>
                    <div className="grow mt-1 ml-4 text-left">
                      <Dialog.Title
                        as="h2"
                        className="text-xl leading-6 font-medium text-gray-900"
                      >
                        {title}
                      </Dialog.Title>
                      <div className="mt-2 mb-2">
                        <p className="w-full sm:w-11/12 text-sm text-foreground-light break-words whitespace-pre-wrap">
                          {message}
                        </p>
                      </div>
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
