#!/usr/bin/env python3
"""Unit tests for call_graph.py module."""

import ast
import json
import os
import tempfile
import unittest

from call_graph import (
    CallGraphVisitor,
    CallSiteExtractor,
    FunctionInfo,
    CallSite,
    CallerInfo,
    CallGraphOutput,
    analyze_file,
    is_test_function,
    is_test_file,
    output_to_json,
    to_dict,
    parse_args,
    build_caller_map,
    analyze_call_graph,
)


class TestIsTestFunction(unittest.TestCase):
    """Tests for is_test_function() function."""

    def test_test_underscore_prefix(self):
        """Test functions starting with test_ are detected."""
        self.assertTrue(is_test_function("test_something"))
        self.assertTrue(is_test_function("test_"))
        self.assertTrue(is_test_function("test_my_function"))

    def test_Test_prefix(self):
        """Test functions starting with Test are detected."""
        self.assertTrue(is_test_function("TestSomething"))
        self.assertTrue(is_test_function("Test"))
        self.assertTrue(is_test_function("TestMyClass"))

    def test_underscore_test_suffix(self):
        """Test functions ending with _test are detected."""
        self.assertTrue(is_test_function("something_test"))
        self.assertTrue(is_test_function("my_function_test"))

    def test_class_method_test_patterns(self):
        """Test class.method patterns are properly handled."""
        self.assertTrue(is_test_function("TestClass.test_method"))
        self.assertTrue(is_test_function("MyClass.test_something"))
        self.assertTrue(is_test_function("SomeClass.TestMethod"))

    def test_non_test_functions(self):
        """Test non-test functions are not detected."""
        self.assertFalse(is_test_function("my_function"))
        self.assertFalse(is_test_function("helper"))
        self.assertFalse(is_test_function("testing"))
        self.assertFalse(is_test_function("attest"))
        self.assertFalse(is_test_function("contest"))
        self.assertFalse(is_test_function("_test_internal"))

    def test_integration_test_prefix(self):
        """Test integration_test pattern is not matched (only test_ prefix)."""
        # Note: integration_test doesn't match test_ prefix
        self.assertFalse(is_test_function("integration_test_something"))

    def test_edge_cases(self):
        """Test edge cases."""
        self.assertFalse(is_test_function(""))
        self.assertTrue(is_test_function("test_"))
        # "_test" ends with "_test", so it matches the suffix pattern
        self.assertTrue(is_test_function("_test"))


class TestIsTestFile(unittest.TestCase):
    """Tests for is_test_file() function."""

    def test_test_prefix_file(self):
        """Test files starting with test_ are detected."""
        self.assertTrue(is_test_file("test_module.py"))
        self.assertTrue(is_test_file("/path/to/test_module.py"))

    def test_test_suffix_file(self):
        """Test files ending with _test.py are detected."""
        self.assertTrue(is_test_file("module_test.py"))
        self.assertTrue(is_test_file("/path/to/module_test.py"))

    def test_tests_directory(self):
        """Test files in tests/ directory are detected."""
        self.assertTrue(is_test_file("tests/module.py"))
        self.assertTrue(is_test_file("/project/tests/module.py"))
        self.assertTrue(is_test_file("project/tests/subdir/module.py"))

    def test_test_directory(self):
        """Test files in test/ directory are detected."""
        self.assertTrue(is_test_file("test/module.py"))
        self.assertTrue(is_test_file("/project/test/module.py"))

    def test_non_test_files(self):
        """Test non-test files are not detected."""
        self.assertFalse(is_test_file("module.py"))
        self.assertFalse(is_test_file("/src/module.py"))
        self.assertFalse(is_test_file("testing.py"))


class TestCallGraphVisitor(unittest.TestCase):
    """Tests for CallGraphVisitor class."""

    def test_simple_function_detection(self):
        """Test detection of simple function definitions."""
        source = """
def hello():
    pass

def world():
    pass
"""
        tree = ast.parse(source)
        visitor = CallGraphVisitor("test.py")
        visitor.visit(tree)

        self.assertEqual(len(visitor.functions), 2)
        names = [fn.name for fn in visitor.functions]
        self.assertIn("hello", names)
        self.assertIn("world", names)

    def test_function_with_calls(self):
        """Test detection of function calls within a function."""
        source = """
def caller():
    helper()
    another_helper()
"""
        tree = ast.parse(source)
        visitor = CallGraphVisitor("test.py")
        visitor.visit(tree)

        self.assertEqual(len(visitor.functions), 1)
        fn = visitor.functions[0]
        self.assertEqual(fn.name, "caller")
        self.assertEqual(len(fn.call_sites), 2)
        targets = [cs.target for cs in fn.call_sites]
        self.assertIn("helper", targets)
        self.assertIn("another_helper", targets)

    def test_method_in_class(self):
        """Test detection of methods in classes."""
        source = """
class MyClass:
    def method_one(self):
        pass

    def method_two(self):
        self.method_one()
"""
        tree = ast.parse(source)
        visitor = CallGraphVisitor("test.py")
        visitor.visit(tree)

        self.assertEqual(len(visitor.functions), 2)
        names = [fn.name for fn in visitor.functions]
        self.assertIn("MyClass.method_one", names)
        self.assertIn("MyClass.method_two", names)

    def test_nested_class(self):
        """Test detection of methods in nested classes."""
        source = """
class Outer:
    def outer_method(self):
        pass

    class Inner:
        def inner_method(self):
            pass
"""
        tree = ast.parse(source)
        visitor = CallGraphVisitor("test.py")
        visitor.visit(tree)

        # Should detect both outer and inner methods
        names = [fn.name for fn in visitor.functions]
        self.assertIn("Outer.outer_method", names)
        # Inner class method should have proper prefix
        self.assertIn("Inner.inner_method", names)

    def test_async_function(self):
        """Test detection of async function definitions."""
        source = """
async def async_func():
    await something()

def sync_func():
    pass
"""
        tree = ast.parse(source)
        visitor = CallGraphVisitor("test.py")
        visitor.visit(tree)

        self.assertEqual(len(visitor.functions), 2)
        names = [fn.name for fn in visitor.functions]
        self.assertIn("async_func", names)
        self.assertIn("sync_func", names)

    def test_function_line_numbers(self):
        """Test that line numbers are correctly captured."""
        source = """def func1():
    pass

def func2():
    pass
"""
        tree = ast.parse(source)
        visitor = CallGraphVisitor("test.py")
        visitor.visit(tree)

        func1 = next(fn for fn in visitor.functions if fn.name == "func1")
        func2 = next(fn for fn in visitor.functions if fn.name == "func2")

        self.assertEqual(func1.line, 1)
        self.assertEqual(func2.line, 4)

    def test_method_call_detection(self):
        """Test detection of method calls (obj.method())."""
        source = """
def caller():
    obj.method()
    result = another.call()
"""
        tree = ast.parse(source)
        visitor = CallGraphVisitor("test.py")
        visitor.visit(tree)

        fn = visitor.functions[0]
        # Should detect method calls
        self.assertTrue(len(fn.call_sites) >= 2)
        # Method calls should be marked as is_method=True
        method_calls = [cs for cs in fn.call_sites if cs.is_method]
        self.assertTrue(len(method_calls) >= 2)

    def test_builtin_calls_filtered(self):
        """Test that builtin calls are filtered out."""
        source = """
def func():
    print("hello")
    len([1, 2, 3])
    my_custom_func()
"""
        tree = ast.parse(source)
        visitor = CallGraphVisitor("test.py")
        visitor.visit(tree)

        fn = visitor.functions[0]
        targets = [cs.target for cs in fn.call_sites]
        # print and len should be filtered
        self.assertNotIn("print", targets)
        self.assertNotIn("len", targets)
        # custom function should be present
        self.assertIn("my_custom_func", targets)

    def test_nested_function_calls_not_double_counted(self):
        """Test that calls in nested functions are not counted in parent."""
        source = """
def outer():
    outer_call()

    def inner():
        inner_call()

    after_inner_call()
"""
        tree = ast.parse(source)
        visitor = CallGraphVisitor("test.py")
        visitor.visit(tree)

        outer_fn = next(fn for fn in visitor.functions if fn.name == "outer")
        inner_fn = next(fn for fn in visitor.functions if fn.name == "inner")

        outer_targets = [cs.target for cs in outer_fn.call_sites]
        inner_targets = [cs.target for cs in inner_fn.call_sites]

        # outer should have its own calls, not inner's
        self.assertIn("outer_call", outer_targets)
        self.assertIn("after_inner_call", outer_targets)
        self.assertNotIn("inner_call", outer_targets)

        # inner should have its own calls
        self.assertIn("inner_call", inner_targets)


class TestCallSiteExtractor(unittest.TestCase):
    """Tests for CallSiteExtractor class."""

    def test_direct_function_call(self):
        """Test extraction of direct function call."""
        source = "my_func()"
        tree = ast.parse(source, mode="eval")
        call_node = tree.body

        extractor = CallSiteExtractor()
        call_site = extractor.extract_call(call_node)

        self.assertIsNotNone(call_site)
        self.assertEqual(call_site.target, "my_func")
        self.assertFalse(call_site.is_method)

    def test_method_call(self):
        """Test extraction of method call."""
        source = "obj.method()"
        tree = ast.parse(source, mode="eval")
        call_node = tree.body

        extractor = CallSiteExtractor()
        call_site = extractor.extract_call(call_node)

        self.assertIsNotNone(call_site)
        self.assertEqual(call_site.target, "obj.method")
        self.assertTrue(call_site.is_method)

    def test_chained_attribute_call(self):
        """Test extraction of chained attribute call."""
        source = "module.submodule.func()"
        tree = ast.parse(source, mode="eval")
        call_node = tree.body

        extractor = CallSiteExtractor()
        call_site = extractor.extract_call(call_node)

        self.assertIsNotNone(call_site)
        self.assertEqual(call_site.target, "module.submodule.func")
        self.assertTrue(call_site.is_method)

    def test_builtin_filtered(self):
        """Test that builtins are filtered out."""
        source = "print('hello')"
        tree = ast.parse(source, mode="eval")
        call_node = tree.body

        extractor = CallSiteExtractor()
        call_site = extractor.extract_call(call_node)

        self.assertIsNone(call_site)


class TestAnalyzeFile(unittest.TestCase):
    """Tests for analyze_file() function."""

    def test_valid_file_analysis(self):
        """Test analysis of a valid Python file."""
        with tempfile.NamedTemporaryFile(
            mode="w", suffix=".py", delete=False
        ) as f:
            f.write("""
def hello():
    pass

def world():
    hello()
""")
            f.flush()
            temp_path = f.name

        try:
            functions = analyze_file(temp_path)
            self.assertEqual(len(functions), 2)
            names = [fn.name for fn in functions]
            self.assertIn("hello", names)
            self.assertIn("world", names)
        finally:
            os.unlink(temp_path)

    def test_syntax_error_handling(self):
        """Test that files with syntax errors return empty list."""
        with tempfile.NamedTemporaryFile(
            mode="w", suffix=".py", delete=False
        ) as f:
            f.write("""
def broken(
    # Missing closing paren and body
""")
            f.flush()
            temp_path = f.name

        try:
            functions = analyze_file(temp_path)
            self.assertEqual(len(functions), 0)
        finally:
            os.unlink(temp_path)

    def test_nonexistent_file(self):
        """Test that nonexistent files return empty list."""
        functions = analyze_file("/nonexistent/path/to/file.py")
        self.assertEqual(len(functions), 0)

    def test_file_with_class(self):
        """Test analysis of file with class definitions."""
        with tempfile.NamedTemporaryFile(
            mode="w", suffix=".py", delete=False
        ) as f:
            f.write("""
class MyClass:
    def method(self):
        pass

def standalone():
    pass
""")
            f.flush()
            temp_path = f.name

        try:
            functions = analyze_file(temp_path)
            self.assertEqual(len(functions), 2)
            names = [fn.name for fn in functions]
            self.assertIn("MyClass.method", names)
            self.assertIn("standalone", names)
        finally:
            os.unlink(temp_path)

    def test_empty_file(self):
        """Test analysis of empty file."""
        with tempfile.NamedTemporaryFile(
            mode="w", suffix=".py", delete=False
        ) as f:
            f.write("")
            f.flush()
            temp_path = f.name

        try:
            functions = analyze_file(temp_path)
            self.assertEqual(len(functions), 0)
        finally:
            os.unlink(temp_path)


class TestOutputToJson(unittest.TestCase):
    """Tests for output_to_json() function."""

    def test_empty_output(self):
        """Test JSON output with no functions."""
        output = CallGraphOutput()
        json_str = output_to_json(output)
        data = json.loads(json_str)

        self.assertIn("functions", data)
        self.assertEqual(len(data["functions"]), 0)
        self.assertNotIn("error", data)

    def test_output_with_functions(self):
        """Test JSON output with functions."""
        fn = FunctionInfo(
            name="my_func",
            file="test.py",
            line=10,
            end_line=20,
            call_sites=[
                CallSite(target="helper", line=15, column=5, is_method=False)
            ],
            called_by=[],
        )
        output = CallGraphOutput(functions=[fn])
        json_str = output_to_json(output)
        data = json.loads(json_str)

        self.assertEqual(len(data["functions"]), 1)
        fn_data = data["functions"][0]
        self.assertEqual(fn_data["name"], "my_func")
        self.assertEqual(fn_data["file"], "test.py")
        self.assertEqual(fn_data["line"], 10)
        self.assertEqual(len(fn_data["call_sites"]), 1)

    def test_output_with_error(self):
        """Test JSON output with error."""
        output = CallGraphOutput(error="Something went wrong")
        json_str = output_to_json(output)
        data = json.loads(json_str)

        self.assertIn("error", data)
        self.assertEqual(data["error"], "Something went wrong")

    def test_output_with_callers(self):
        """Test JSON output with caller information."""
        fn = FunctionInfo(
            name="callee",
            file="test.py",
            line=10,
            end_line=15,
            call_sites=[],
            called_by=[
                CallerInfo(function="caller", file="test.py", line=25)
            ],
        )
        output = CallGraphOutput(functions=[fn])
        json_str = output_to_json(output)
        data = json.loads(json_str)

        fn_data = data["functions"][0]
        self.assertEqual(len(fn_data["called_by"]), 1)
        caller = fn_data["called_by"][0]
        self.assertEqual(caller["function"], "caller")


class TestToDict(unittest.TestCase):
    """Tests for to_dict() helper function."""

    def test_simple_dataclass(self):
        """Test conversion of simple dataclass."""
        call_site = CallSite(
            target="func", line=10, column=5, is_method=False
        )
        result = to_dict(call_site)

        self.assertEqual(result["target"], "func")
        self.assertEqual(result["line"], 10)
        self.assertEqual(result["column"], 5)
        self.assertEqual(result["is_method"], False)

    def test_nested_dataclass(self):
        """Test conversion of nested dataclass."""
        fn = FunctionInfo(
            name="func",
            file="test.py",
            line=1,
            end_line=5,
            call_sites=[
                CallSite(target="helper", line=3, column=4, is_method=False)
            ],
            called_by=[],
        )
        result = to_dict(fn)

        self.assertEqual(result["name"], "func")
        self.assertEqual(len(result["call_sites"]), 1)
        self.assertEqual(result["call_sites"][0]["target"], "helper")


class TestParseArgs(unittest.TestCase):
    """Tests for parse_args() function."""

    def test_files_only(self):
        """Test parsing with only file arguments."""
        files, functions = parse_args(["file1.py", "file2.py"])

        self.assertEqual(files, ["file1.py", "file2.py"])
        self.assertEqual(functions, [])

    def test_with_functions_flag(self):
        """Test parsing with --functions flag."""
        files, functions = parse_args(
            ["file.py", "--functions", "func1,func2"]
        )

        self.assertEqual(files, ["file.py"])
        self.assertEqual(functions, ["func1", "func2"])

    def test_functions_with_spaces(self):
        """Test parsing functions with spaces around commas."""
        _files, functions = parse_args(
            ["file.py", "--functions", "func1, func2, func3"]
        )

        self.assertEqual(functions, ["func1", "func2", "func3"])

    def test_empty_args(self):
        """Test parsing empty arguments."""
        files, functions = parse_args([])

        self.assertEqual(files, [])
        self.assertEqual(functions, [])


class TestBuildCallerMap(unittest.TestCase):
    """Tests for build_caller_map() function."""

    def test_simple_caller_relationship(self):
        """Test building caller map with simple relationships."""
        caller = FunctionInfo(
            name="caller",
            file="test.py",
            line=1,
            end_line=5,
            call_sites=[
                CallSite(target="callee", line=3, column=4, is_method=False)
            ],
            called_by=[],
        )
        callee = FunctionInfo(
            name="callee",
            file="test.py",
            line=10,
            end_line=15,
            call_sites=[],
            called_by=[],
        )

        caller_map = build_caller_map([caller, callee])

        self.assertIn("callee", caller_map)
        self.assertEqual(len(caller_map["callee"]), 1)
        self.assertEqual(caller_map["callee"][0].function, "caller")

    def test_no_callers(self):
        """Test function with no callers."""
        fn = FunctionInfo(
            name="isolated",
            file="test.py",
            line=1,
            end_line=5,
            call_sites=[],
            called_by=[],
        )

        caller_map = build_caller_map([fn])

        self.assertIn("isolated", caller_map)
        self.assertEqual(len(caller_map["isolated"]), 0)


class TestAnalyzeCallGraph(unittest.TestCase):
    """Tests for analyze_call_graph() function."""

    def test_empty_files_list(self):
        """Test with empty files list."""
        result = analyze_call_graph([])
        self.assertEqual(len(result.functions), 0)
        self.assertIsNone(result.error)

    def test_nonexistent_files(self):
        """Test with nonexistent files."""
        result = analyze_call_graph(["/nonexistent/file.py"])
        self.assertEqual(len(result.functions), 0)
        self.assertIsNotNone(result.error)

    def test_with_target_functions(self):
        """Test filtering by target functions."""
        with tempfile.NamedTemporaryFile(
            mode="w", suffix=".py", delete=False
        ) as f:
            f.write("""
def target_func():
    pass

def other_func():
    pass
""")
            f.flush()
            temp_path = f.name

        try:
            result = analyze_call_graph(
                [temp_path], target_functions=["target_func"]
            )
            self.assertEqual(len(result.functions), 1)
            self.assertEqual(result.functions[0].name, "target_func")
        finally:
            os.unlink(temp_path)


if __name__ == "__main__":
    unittest.main()
