from ...ffi.ffi import lib


class CTextFormatter:
    def __init__(self, style_id: int = 0, max_depth:int=3):
        self._id = lib.NewTextFormatter(style_id, max_depth)
