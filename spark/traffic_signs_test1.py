from pyspark.sql import SparkSession
from pyspark.sql.functions import col, count

def filter_by_pattern(input_file, output_file, pattern):
    """
    Task 1: Filter all lines that contain the given pattern in any column
    and return only OBJECTID and Sign_Type to output file.
    """
    # Initialize Spark Session
    spark = SparkSession.builder \
        .appName("Filter By Pattern") \
        .getOrCreate()

    # Load the dataset
    df = spark.read.csv(input_file, header=True, inferSchema=True)

    # Filter rows where any column contains the pattern
    filtered_df = df.filter(
        df.columns[1].contains(pattern)
    ).select("OBJECTID", "Sign_Type")

    # Save results to output file
    filtered_df.write.csv(output_file, header=True)

    spark.stop()

def count_categories_by_sign_post(input_file, output_file, sign_post_type):
    """
    Task 2: Among traffic signs having 'Sign_Post' of type `sign_post_type` (exact match),
    return the count of 'Category'.
    """
    # Initialize Spark Session
    spark = SparkSession.builder \
        .appName("Count Categories by Sign Post") \
        .getOrCreate()

    # Load the dataset
    df = spark.read.csv(input_file, header=True, inferSchema=True)

    # Filter rows where 'Sign_Post' matches sign_post_type and group by 'Category'
    result_df = df.filter(col("Sign_Post") == sign_post_type) \
                  .groupBy("Category") \
                  .agg(count("Category").alias("Category_Count"))

    # Save results to output file
    result_df.write.csv(output_file, header=True)

    spark.stop()

if __name__ == "__main__":
    # Example usage for Task 1
    # TA provides pattern X = "No Outlet"
    pattern_to_search = "No Outlet"
    filter_by_pattern("TrafficSigns_1000.csv", "filtered_output_spark.csv", pattern_to_search)
