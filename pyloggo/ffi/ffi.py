import ctypes
import os

# Загрузка loggo.so
lib_path = os.path.join(os.path.dirname(__file__), "loggo.so")
lib = ctypes.CDLL(lib_path)

# Тип указателя из Go
uintptr_t = ctypes.c_ulong

# Объявление сигнатур
lib.NewLoggerWithSingleRoute.argtypes = [uintptr_t]
lib.NewLoggerWithSingleRoute.restype = uintptr_t
lib.NewLoggerWithRoutes.argtypes = [ctypes.POINTER(uintptr_t), ctypes.c_int]
lib.NewLoggerWithRoutes.restype = uintptr_t

for level in ["trace", "debug", "info", "warning", "error", "exception"]:
    fn = getattr(lib, f"Logger_{level.capitalize()}ToRoute")
    fn.argtypes = [ctypes.c_ulong, ctypes.c_char_p, ctypes.c_char_p]
    fn.restype = None


lib.FreeLogger.argtypes = [uintptr_t]
lib.FreeLogger.restype = None
lib.NewFormatStyle.argtypes = [
    ctypes.c_int,
    ctypes.c_int,
    ctypes.c_int,
    ctypes.c_char_p,
    ctypes.c_char_p,
    ctypes.c_char_p,
]
lib.NewFormatStyle.restype = ctypes.c_ulong
lib.NewTextFormatter.argtypes = [ctypes.c_ulong, ctypes.c_int]
lib.NewTextFormatter.restype = ctypes.c_ulong
lib.NewStdoutWriter.argtypes = []
lib.NewStdoutWriter.restype = ctypes.c_ulong
lib.NewJsonFormatter.argtypes = [ctypes.c_ulong, ctypes.c_int]
lib.NewJsonFormatter.restype = ctypes.c_ulong
lib.NewFileWriter.argtypes = [
    ctypes.c_char_p,  # path
    ctypes.c_long,  # maxSizeMB
    ctypes.c_int,  # maxBackups
    ctypes.c_char_p,  # interval
    ctypes.c_char_p,  # compress
]
lib.NewFileWriter.restype = ctypes.c_ulong
