from .ffi.ffi import lib
from .json import _serialize_fields
from .route import RouteProcessor
from .c import CLogger
import sys
import linecache
import os
from .enums import LogLevel
from typing import Any

import threading


class Logger:
    def __init__(self, routes: list[RouteProcessor]) -> None:
        route_ids = [r.id for r in routes]
        self._c_logger = CLogger(route_ids)
        self._routes = routes

    @property
    def id(self) -> int:
        return self._c_logger._id

    def _log(self, method: str, msg: str, **kwargs) -> None:
        level: int = getattr(LogLevel, method.capitalize())

        msg_b = msg.encode()
        for route in self._routes:
            route_fields = self._resolve_fields(route, level, kwargs)
            fields_b = _serialize_fields(route_fields)
            getattr(lib, f"Logger_{method.capitalize()}ToRoute")(
                route.id, msg_b, fields_b
            )

    def _resolve_fields(
        self,
        route: RouteProcessor,
        level: LogLevel,
        fields: dict[str, Any],
    ) -> dict[str, Any]:
        fields_cp = dict(fields)
        tb = False
        if route.tb and route.tb_level <= level:
            fields_cp["tb"] = self._add_traceback(max_depth=route.tb_max_depth)
            tb = True
        if not tb and route.scope:
            fields_cp["scope"] = self._add_scope()

        return fields_cp

    @staticmethod
    def _add_scope(frame_depth: int = 5) -> str:
        try:
            frame = sys._getframe(frame_depth)
            filename = os.path.basename(frame.f_code.co_filename)
            lineno = frame.f_lineno
            func = frame.f_code.co_name
            return f"{filename}:{lineno} in {func}()"
        except Exception:
            return "<scope unavailable>"

    @staticmethod
    def _add_traceback(max_depth: int = 10, skip: int = 5) -> str:
        lines = []
        frame = sys._getframe(skip)

        for _ in range(max_depth):
            if frame is None:
                break

            filename_full = frame.f_code.co_filename
            filename = os.path.basename(filename_full)
            lineno = frame.f_lineno
            func = frame.f_code.co_name

            code_line = linecache.getline(filename_full, lineno).strip()

            lines.append(
                f'  File "{filename}", line {lineno}, in {func}()\n    {code_line}\n'
            )

            frame = frame.f_back

        return "".join(lines)

    def trace(self, msg: str, **kwargs):
        self._log("trace", msg, **kwargs)

    def debug(self, msg: str, **kwargs):
        self._log("debug", msg, **kwargs)

    def info(self, msg: str, **kwargs):
        self._log("info", msg, **kwargs)

    def warning(self, msg: str, **kwargs):
        self._log("warning", msg, **kwargs)

    def error(self, msg: str, **kwargs):
        self._log("error", msg, **kwargs)

    def exception(self, msg: str, **kwargs):
        self._log("exception", msg, **kwargs)

    def close(self):
        self._c_logger.close()

    def __del__(self):
        try:
            self.close()
        except Exception:
            pass


def create_default_logger() -> Logger:
    router = RouteProcessor()
    return Logger([router])


class GlobalLogger:
    def __init__(self):
        self._lock = threading.Lock()
        self._logger = self._create_default_logger()

    def _create_default_logger(self) -> Logger:
        return create_default_logger()

    def info(self, msg, **kwargs):
        self._logger.info(msg, **kwargs)

    def debug(self, msg, **kwargs):
        self._logger.debug(msg, **kwargs)

    def warning(self, msg, **kwargs):
        self._logger.warning(msg, **kwargs)

    def error(self, msg, **kwargs):
        self._logger.error(msg, **kwargs)

    def exception(self, msg, **kwargs):
        self._logger.exception(msg, **kwargs)

    def trace(self, msg, **kwargs):
        self._logger.trace(msg, **kwargs)

    def add(self, route: RouteProcessor):
        with self._lock:
            # пересоздаём logger с новым роутом
            self._logger.close()
            self._logger = Logger(routes=[route])

    def remove(self):
        with self._lock:
            self._logger.close()
            self._logger = self._create_default_logger()

    def configure(self, routes: list):
        with self._lock:
            self._logger.close()
            self._logger = Logger(routes=list(routes))

    def close(self):
        with self._lock:
            self._logger.close()

    def __del__(self):
        try:
            self._logger.close()
        except Exception:
            pass
