"""
An example of using the llama_index library to build
a question answering system for major league cricket data.
"""
import sys
import os
import logging
import json

from llama_index import ServiceContext, StorageContext, SQLDatabase, VectorStoreIndex, schema
from llama_index.indices.struct_store import SQLTableRetrieverQueryEngine
from llama_index.objects import SQLTableNodeMapping, ObjectIndex, SQLTableSchema
from llama_index.tools.query_engine import QueryEngineTool
from llama_index.embeddings import OpenAIEmbedding
from llama_index.vector_stores import ChromaVectorStore
from llama_index.vector_stores.types import VectorStoreInfo
from llama_index.indices.vector_store.retrievers import VectorIndexAutoRetriever
from llama_index.query_engine.retriever_query_engine import RetrieverQueryEngine
from llama_index.query_engine import SQLAutoVectorQueryEngine
from llama_index.llms import OpenAI
from sqlalchemy import create_engine, MetaData

__import__('pysqlite3')
sys.modules['sqlite3'] = sys.modules.pop('pysqlite3')
import chromadb

import openai
import gradio as gr

logging.basicConfig(stream=sys.stdout, level=logging.INFO)
logging.getLogger().addHandler(logging.StreamHandler(stream=sys.stdout))

openai.api_key = os.environ["OPENAI_API_KEY"]
embed_model = OpenAIEmbedding()
llm = OpenAI(model="gpt-4")
service_context = ServiceContext.from_defaults(embed_model=embed_model, llm=llm)

def get_table_names(engine):
    """Get all table names from a database."""
    metadata = MetaData()
    metadata.reflect(bind=engine)
    table_names = []
    for table in metadata.tables.values():
        table_names.append(table.name)
    return table_names

def get_sql_tool():
    """Get a tool for translating natural language queries into SQL queries."""
    engine = create_engine('sqlite:///data/player_stats.db')
    sql_database = SQLDatabase(engine=engine)
    table_node_mapping = SQLTableNodeMapping(sql_database)
    all_table_names = get_table_names(engine)
    table_schema_objs = []
    for table_name in all_table_names:
        table_schema_objs.append(SQLTableSchema(table_name=table_name))
    obj_index = ObjectIndex.from_objects(
        table_schema_objs,
        table_node_mapping,
        VectorStoreIndex,
    )
    sql_query_engine = SQLTableRetrieverQueryEngine(
    sql_database,
    obj_index.as_retriever(similarity_top_k=1),
    )
    st = QueryEngineTool.from_defaults(
        query_engine=sql_query_engine,
        description=(
            "Useful for translating a natural language query into a SQL query over"
            " a database named player_stats. The database contains 4 tables:"
            " batting_players, bowling_players, teams, and matches. batting_players table contains"
            " batting related information about each player, bowling_players table contains"
            " bowling related information about each player, teams table contains information about each"
            " team, and matches table contains information about each match." 
        ),
    )
    return st

def get_index():
    """Get a vector store index for the chroma database."""
    db = chromadb.PersistentClient(path="./chroma_db")
    try:
        chroma_collection = db.get_collection("mlc_articles")
    except ValueError:
        return None
    vector_store = ChromaVectorStore(chroma_collection=chroma_collection)
    index = VectorStoreIndex.from_vector_store(
        vector_store, service_context=service_context
    )
    return index

def build_index():
    """Build a vector store index for the chroma database."""
    with open("data/full-articles.json", encoding="utf-8") as f:
        articles = json.load(f)

    documents = []
    num_empty_articles = 0
    for article in articles:
        b = article['body']
        if len(b) > 0:
            d = schema.Document(text=b)
            documents.append(d)
        else:
            num_empty_articles += 1
    print(f"Number of empty articles: {num_empty_articles}")

    chroma_client = chromadb.PersistentClient(path="./chroma_db")
    chroma_collection = chroma_client.create_collection("mlc_articles")
    vector_store = ChromaVectorStore(chroma_collection=chroma_collection)
    storage_context = StorageContext.from_defaults(vector_store=vector_store)
    index = VectorStoreIndex.from_documents(
        documents, storage_context=storage_context, service_context=service_context
    )
    return index

def get_vector_tool():
    """Get a tool for translating natural language queries into vector queries."""
    index = get_index()
    if index is None:
        print("Building index...")
        index = build_index()

    vector_store_info = VectorStoreInfo(
      content_info="articles about major league cricket or MLC",
        metadata_info=[]
    )
    vector_auto_retriever = VectorIndexAutoRetriever(
        index, vector_store_info=vector_store_info
    )
    retriever_query_engine = RetrieverQueryEngine.from_args(
        vector_auto_retriever, service_context=service_context
    )
    vt = QueryEngineTool.from_defaults(
       query_engine=retriever_query_engine,
        description=(
        "Useful for answering semantic questions about major league cricket or MLC"
        ),
    )
    return vt

if __name__ == '__main__':
    sql_tool = get_sql_tool()
    vector_tool = get_vector_tool()
    query_engine = SQLAutoVectorQueryEngine(
    sql_tool, vector_tool, service_context=service_context
    )

    def handle_query(query):
        """Handle the query."""
        if len(query) > 200:
            response = "Sorry, your query is too long. Please try again with a shorter query."
            return response
        try:
            response = query_engine.query(query)
        except Exception as e: # pylint: disable=broad-exception-caught
            print(e)
            response = "Sorry, an error ocurred while answering the query. Please try again later."
        return response

    demo = gr.Interface(
        fn=handle_query,
        inputs=gr.Textbox(lines=2, placeholder="Enter query here..."),
        outputs=gr.Textbox(),
        title="MLC Guru",
        description="Ask questions about Major League Cricket (MLC)",
        examples=[
        ["How many teams are in MLC and what are they?"],
        ["What is the name of the player with the best batting strike rate and what is that value?"],
        ["Which Australian states are taking part in Major League Cricket and what are they doing?"],
        ["Who is the owner of Major league cricket and how much does it cost to run the league?"],
        ["What are the names of teams that won matches played in Morrisville?"]
        ],
        cache_examples=True,
    )

    if "PORT" in os.environ:
        port = int(os.environ["PORT"])
        demo.launch(share=False, server_name="0.0.0.0", server_port=port)
    else:
        demo.launch(share=False)
    
    # while True:
    #     q = input("Enter query (or quit): ")
    #     if q.lower() == "quit":
    #         break
    #     response = query_engine.query(q)
    #     print(response)
    #     print(response.metadata)
