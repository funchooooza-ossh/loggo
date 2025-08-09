from ...ffi.ffi import lib


class CJsonFormatter:
    def __init__(self, style_id: int = 0, max_depth:int=3):
        self._id = lib.NewJsonFormatter(style_id, max_depth)
