export interface FileChunk {
  file_id: string;           // Unique identifier for the original file
  chunk_index: number;        // Index of this chunk (0-based)
  total_chunks: number;       // Total number of chunks for this file
  chunk_data: Blob;           // The actual chunk data
  original_filename: string;  // Name of the original file
  file_size: number;          // Total size of the original file
  chunk_size: number;         // Size of this specific chunk
}

export interface ChunkedFile {
  file_id: string;
  original_file: File;
  chunks: FileChunk[];
  total_chunks: number;
}

