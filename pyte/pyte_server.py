#!/usr/bin/env python3
import sys
import json
import pyte

class ScreenServer:
    def __init__(self):
        self.screen = None
        
    def handle_command(self, cmd):
        method = cmd["method"]
        args = cmd.get("args", [])
        kwargs = cmd.get("kwargs", {})
        
        # Special handling for init
        if method == "__init__":
            if len(args) >= 2:
                cols, lines = int(args[0]), int(args[1])
                self.screen = pyte.Screen(cols, lines)
                return {"success": True}
            else:
                return {"success": False, "error": "Need columns and lines for init"}
        
        if self.screen is None:
            return {"success": False, "error": "Screen not initialized"}
        
        # Handle the method call
        method_name = method
        
        # Map Go method names to Python method names
        method_map = {
            "Draw": "draw",
            "Bell": "bell",
            "Backspace": "backspace",
            "Tab": "tab",
            "Linefeed": "linefeed",
            "CarriageReturn": "carriage_return",
            "ShiftOut": "shift_out",
            "ShiftIn": "shift_in",
            "CursorUp": "cursor_up",
            "CursorDown": "cursor_down",
            "CursorForward": "cursor_forward",
            "CursorBack": "cursor_back",
            "CursorUp1": "cursor_up1",
            "CursorDown1": "cursor_down1",
            "CursorPosition": "cursor_position",
            "CursorToColumn": "cursor_to_column",
            "CursorToLine": "cursor_to_line",
            "Reset": "reset",
            "Index": "index",
            "ReverseIndex": "reverse_index",
            "SetTabStop": "set_tab_stop",
            "ClearTabStop": "clear_tab_stop",
            "SaveCursor": "save_cursor",
            "RestoreCursor": "restore_cursor",
            "InsertLines": "insert_lines",
            "DeleteLines": "delete_lines",
            "InsertCharacters": "insert_characters",
            "DeleteCharacters": "delete_characters",
            "EraseCharacters": "erase_characters",
            "EraseInLine": "erase_in_line",
            "EraseInDisplay": "erase_in_display",
            "SetMode": "set_mode",
            "ResetMode": "reset_mode",
            "SelectGraphicRendition": "select_graphic_rendition",
            "DefineCharset": "define_charset",
            "SetMargins": "set_margins",
            "ReportDeviceAttributes": "report_device_attributes",
            "ReportDeviceStatus": "report_device_status",
            "SetTitle": "set_title",
            "SetIconName": "set_icon_name",
            "AlignmentDisplay": "alignment_display",
            "Debug": "debug",
            "WriteProcessInput": "write_process_input",
        }
        
        python_method = method_map.get(method, method)
        
        if hasattr(self.screen, python_method):
            try:
                # Handle None arguments
                if args and args[0] is None:
                    args = []
                    
                result = getattr(self.screen, python_method)(*args, **kwargs)
                
                if python_method == "display":
                    return {"success": True, "result": result}
                else:
                    return {"success": True}
            except Exception as e:
                import traceback
                return {"success": False, "error": str(e), "trace": traceback.format_exc()}
        else:
            return {"success": False, "error": f"Unknown method: {python_method}"}
    
    def run(self):
        sys.stderr.write("Python screen server started\n")
        sys.stderr.flush()
        
        for line in sys.stdin:
            try:
                cmd = json.loads(line.strip())
                response = self.handle_command(cmd)
                print(json.dumps(response), flush=True)
            except Exception as e:
                import traceback
                error_response = {
                    "success": False, 
                    "error": str(e),
                    "trace": traceback.format_exc()
                }
                print(json.dumps(error_response), flush=True)

if __name__ == "__main__":
    server = ScreenServer()
    server.run()