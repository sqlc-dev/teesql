using System.Reflection;
using System.Text.Json;
using System.Text.Json.Serialization;
using Microsoft.SqlServer.TransactSql.ScriptDom;

class Program
{
    static int Main(string[] args)
    {
        if (args.Length < 1)
        {
            Console.Error.WriteLine("Usage: TsqlAstParser <sql-file> [output-json-file]");
            return 1;
        }

        string sqlFile = args[0];
        string outputFile = args.Length > 1 ? args[1] : "ast.json";

        if (!File.Exists(sqlFile))
        {
            Console.Error.WriteLine($"Error: SQL file '{sqlFile}' not found.");
            return 1;
        }

        string sql = File.ReadAllText(sqlFile);

        var parser = new TSql160Parser(initialQuotedIdentifiers: true);

        using var reader = new StringReader(sql);
        var fragment = parser.Parse(reader, out var errors);

        if (errors.Count > 0)
        {
            Console.Error.WriteLine("Parse errors:");
            foreach (var error in errors)
            {
                Console.Error.WriteLine($"  Line {error.Line}, Column {error.Column}: {error.Message}");
            }
            return 1;
        }

        var astConverter = new AstToJsonConverter();
        var jsonObject = astConverter.Convert(fragment);

        var options = new JsonSerializerOptions
        {
            WriteIndented = true,
            DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull
        };

        string json = JsonSerializer.Serialize(jsonObject, options);
        File.WriteAllText(outputFile, json);

        Console.WriteLine($"AST written to {outputFile}");
        return 0;
    }
}

/// <summary>
/// Converts TSqlFragment AST nodes to JSON-serializable dictionaries
/// </summary>
class AstToJsonConverter
{
    private readonly HashSet<object> _visited = new();

    // Properties to skip during serialization (internal/infrastructure properties)
    private static readonly HashSet<string> SkipProperties = new()
    {
        "ScriptTokenStream",
        "FirstTokenIndex",
        "LastTokenIndex",
        "FragmentLength",
        "StartOffset",
        "StartLine",
        "StartColumn"
    };

    public Dictionary<string, object?> Convert(TSqlFragment fragment)
    {
        _visited.Clear();
        return ConvertNode(fragment);
    }

    private Dictionary<string, object?> ConvertNode(TSqlFragment node)
    {
        if (_visited.Contains(node))
        {
            return new Dictionary<string, object?> { ["$ref"] = node.GetType().Name };
        }
        _visited.Add(node);

        var result = new Dictionary<string, object?>
        {
            ["$type"] = node.GetType().Name
        };

        var type = node.GetType();
        var properties = type.GetProperties(BindingFlags.Public | BindingFlags.Instance);

        foreach (var prop in properties)
        {
            if (SkipProperties.Contains(prop.Name))
                continue;

            // Skip indexers
            if (prop.GetIndexParameters().Length > 0)
                continue;

            try
            {
                var value = prop.GetValue(node);
                var convertedValue = ConvertValue(value);

                if (convertedValue != null)
                {
                    result[prop.Name] = convertedValue;
                }
            }
            catch
            {
                // Skip properties that throw exceptions
            }
        }

        return result;
    }

    private object? ConvertValue(object? value)
    {
        if (value == null)
            return null;

        var type = value.GetType();

        // Handle TSqlFragment nodes
        if (value is TSqlFragment fragment)
        {
            return ConvertNode(fragment);
        }

        // Handle collections of TSqlFragment
        if (value is System.Collections.IEnumerable enumerable && type != typeof(string))
        {
            var list = new List<object?>();
            foreach (var item in enumerable)
            {
                var converted = ConvertValue(item);
                if (converted != null)
                {
                    list.Add(converted);
                }
            }
            return list.Count > 0 ? list : null;
        }

        // Handle primitive types and enums
        if (type.IsPrimitive || type.IsEnum || value is string || value is decimal)
        {
            if (type.IsEnum)
            {
                return value.ToString();
            }
            return value;
        }

        // For other complex types, just return the string representation
        return value.ToString();
    }
}
